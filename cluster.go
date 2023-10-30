package agscheduler

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/rpc"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Node struct {
	Id                string
	MainEndpoint      string
	Endpoint          string
	SchedulerEndpoint string
	SchedulerQueue    string
	queueMap          map[string]map[string]map[string]any
}

func (n *Node) toClusterNode() *ClusterNode {
	return &ClusterNode{
		Id:                n.Id,
		MainEndpoint:      n.MainEndpoint,
		Endpoint:          n.Endpoint,
		SchedulerEndpoint: n.SchedulerEndpoint,
		SchedulerQueue:    n.SchedulerQueue,
		queueMap:          n.queueMap,
	}
}

type ClusterNode struct {
	Id                string
	MainEndpoint      string
	Endpoint          string
	SchedulerEndpoint string
	SchedulerQueue    string
	queueMap          map[string]map[string]map[string]any
}

func (cn *ClusterNode) toNode() *Node {
	return &Node{
		Id:                cn.Id,
		MainEndpoint:      cn.MainEndpoint,
		Endpoint:          cn.Endpoint,
		SchedulerEndpoint: cn.SchedulerEndpoint,
		SchedulerQueue:    cn.SchedulerQueue,
		queueMap:          cn.queueMap,
	}
}

func (cn *ClusterNode) QueueMap() map[string]map[string]map[string]any {
	return cn.queueMap
}

func (cn *ClusterNode) setId() {
	cn.Id = strings.Replace(uuid.New().String(), "-", "", -1)[:16]
}

func (cn *ClusterNode) init(ctx context.Context) error {
	cn.setId()
	cn.registerNode(cn)

	go cn.checkNode(ctx)

	return nil
}

func (cn *ClusterNode) registerNode(n *ClusterNode) {
	if cn.queueMap == nil {
		cn.queueMap = make(map[string]map[string]map[string]any)
	}
	if _, ok := cn.queueMap[n.SchedulerQueue]; !ok {
		cn.queueMap[n.SchedulerQueue] = map[string]map[string]any{}
	}
	cn.queueMap[n.SchedulerQueue][n.Id] = map[string]any{
		"id":                 n.Id,
		"main_endpoint":      n.MainEndpoint,
		"endpoint":           n.Endpoint,
		"scheduler_endpoint": n.SchedulerEndpoint,
		"scheduler_queue":    n.SchedulerQueue,
		"health":             true,
		"last_register_time": time.Now().UTC(),
	}
}

func (cn *ClusterNode) choiceNode(queue string) (*ClusterNode, error) {
	cns := make([]*ClusterNode, 0)
	for q, v := range cn.queueMap {
		if queue != "" && q != queue {
			continue
		}
		for id, v2 := range v {
			if !v2["health"].(bool) {
				continue
			}
			cns = append(cns, &ClusterNode{
				Id:                id,
				MainEndpoint:      v2["main_endpoint"].(string),
				Endpoint:          v2["endpoint"].(string),
				SchedulerEndpoint: v2["scheduler_endpoint"].(string),
				SchedulerQueue:    v2["scheduler_queue"].(string),
			})
		}
	}

	cns_count := len(cns)
	if cns_count != 0 {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		i := rand.Intn(cns_count)
		return cns[i], nil
	}

	return &ClusterNode{}, fmt.Errorf("cluster node not found")
}

func (cn *ClusterNode) checkNode(ctx context.Context) {
	interval := 200 * time.Millisecond
	timer := time.NewTimer(interval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			now := time.Now().UTC()
			for _, v := range cn.queueMap {
				for id, v2 := range v {
					if cn.Id == id {
						continue
					}
					endpoint := v2["endpoint"].(string)
					lastRegisterTime := v2["last_register_time"].(time.Time)
					if now.Sub(lastRegisterTime) > 1*time.Second {
						delete(v, id)
						slog.Warn(fmt.Sprintf("Cluster node `%s:%s` is deleted", id, endpoint))
					} else if now.Sub(lastRegisterTime) > 200*time.Millisecond {
						v2["health"] = false
						slog.Warn(fmt.Sprintf("Cluster node `%s:%s` is unhealthy", id, endpoint))
					}
				}
			}
			timer.Reset(interval)
		}
	}
}

func (cn *ClusterNode) RPCRegister(args *Node, reply *Node) {
	slog.Info(fmt.Sprintf("Registration from the cluster node `%s:%s`:", args.Id, args.Endpoint))
	slog.Info(fmt.Sprintf("Cluster Node Scheduler RPC Service listening at: %s", args.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Node Scheduler RPC Service queue: `%s`", args.SchedulerQueue))

	cn.registerNode(args.toClusterNode())

	reply.Id = cn.Id
	reply.MainEndpoint = cn.MainEndpoint
	reply.Endpoint = cn.Endpoint
	reply.SchedulerEndpoint = cn.SchedulerEndpoint
	reply.SchedulerQueue = cn.SchedulerQueue
}

func (cn *ClusterNode) RPCPing(args *Node, reply *Node) {
	cn.registerNode(args.toClusterNode())
}

func (cn *ClusterNode) RegisterNodeRemote(ctx context.Context) error {
	slog.Info(fmt.Sprintf("Register with cluster main `%s`:", cn.MainEndpoint))

	rClient, err := rpc.DialHTTP("tcp", cn.MainEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster main: `%s`, error: %s", cn.MainEndpoint, err)
	}

	var main Node
	c := make(chan error, 1)
	go func() { c <- rClient.Call("CRPCService.Register", cn.toNode(), &main) }()
	select {
	case err := <-c:
		if err != nil {
			return fmt.Errorf("failed to register to cluster main, error: %s", err)
		}
	case <-time.After(3 * time.Second):
		return fmt.Errorf("register to cluster main timeout: %s", err)
	}

	slog.Info(fmt.Sprintf("Cluster Main Scheduler RPC Service listening at: %s", main.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Main Scheduler RPC Service queue: `%s`", main.SchedulerQueue))

	go cn.heartbeatRemote(ctx)

	return nil
}

func (cn *ClusterNode) heartbeatRemote(ctx context.Context) {
	interval := 100 * time.Millisecond
	timer := time.NewTimer(interval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if err := cn.pingRemote(); err != nil {
				slog.Info(fmt.Sprintf("Ping remote error: %s", err))
				timer.Reset(time.Second)
			} else {
				timer.Reset(interval)
			}
		}
	}
}

func (cn *ClusterNode) pingRemote() error {
	rClient, err := rpc.DialHTTP("tcp", cn.MainEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster main: `%s`, error: %s", cn.MainEndpoint, err)
	}

	var main Node
	c := make(chan error, 1)
	go func() { c <- rClient.Call("CRPCService.Ping", cn.toNode(), &main) }()
	select {
	case err := <-c:
		if err != nil {
			return fmt.Errorf("failed to ping to cluster main, error: %s", err)
		}
	case <-time.After(200 * time.Millisecond):
		return fmt.Errorf("ping to cluster main timeout: %s", err)
	}

	return nil
}
