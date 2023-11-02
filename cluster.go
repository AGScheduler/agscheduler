package agscheduler

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/rpc"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Node struct {
	Id                string
	MainEndpoint      string
	Endpoint          string
	EndpointHTTP      string
	SchedulerEndpoint string
	Queue             string
	QueueMap          map[string]map[string]map[string]any
}

func (n *Node) toClusterNode() *ClusterNode {
	return &ClusterNode{
		Id:                n.Id,
		MainEndpoint:      n.MainEndpoint,
		Endpoint:          n.Endpoint,
		EndpointHTTP:      n.EndpointHTTP,
		SchedulerEndpoint: n.SchedulerEndpoint,
		Queue:             n.Queue,

		queueMap: n.QueueMap,
	}
}

type ClusterNode struct {
	Id                string
	MainEndpoint      string
	Endpoint          string
	EndpointHTTP      string
	SchedulerEndpoint string
	Queue             string

	queueMap map[string]map[string]map[string]any
}

func (cn *ClusterNode) toNode() *Node {
	return &Node{
		Id:                cn.Id,
		MainEndpoint:      cn.MainEndpoint,
		Endpoint:          cn.Endpoint,
		EndpointHTTP:      cn.EndpointHTTP,
		SchedulerEndpoint: cn.SchedulerEndpoint,
		Queue:             cn.Queue,
		QueueMap:          cn.queueMap,
	}
}

func (cn *ClusterNode) setQueueMap(qmap map[string]map[string]map[string]any) {
	var mutex sync.Mutex

	mutex.Lock()
	cn.queueMap = qmap
	mutex.Unlock()
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

	if cn.MainEndpoint == cn.Endpoint {
		go cn.checkNode(ctx)
	}

	return nil
}

func (cn *ClusterNode) registerNode(n *ClusterNode) {
	var mutex sync.Mutex

	mutex.Lock()

	if cn.queueMap == nil {
		cn.queueMap = make(map[string]map[string]map[string]any)
	}
	if _, ok := cn.queueMap[n.Queue]; !ok {
		cn.queueMap[n.Queue] = map[string]map[string]any{}
	}
	cn.queueMap[n.Queue][n.Id] = map[string]any{
		"id":                 n.Id,
		"main_endpoint":      n.MainEndpoint,
		"endpoint":           n.Endpoint,
		"endpoint_http":      n.EndpointHTTP,
		"scheduler_endpoint": n.SchedulerEndpoint,
		"queue":              n.Queue,
		"health":             true,
		"last_register_time": time.Now().UTC(),
	}

	mutex.Unlock()
}

func (cn *ClusterNode) choiceNode(queues []string) (*ClusterNode, error) {
	cns := make([]*ClusterNode, 0)
	for q, v := range cn.queueMap {
		if len(queues) != 0 && !slices.Contains(queues, q) {
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
				EndpointHTTP:      v2["endpoint_http"].(string),
				SchedulerEndpoint: v2["scheduler_endpoint"].(string),
				Queue:             v2["queue"].(string),
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
					if now.Sub(lastRegisterTime) > 5*time.Minute {
						delete(v, id)
						slog.Warn(fmt.Sprintf("Cluster node `%s:%s` have been deleted because unhealthy", id, endpoint))
					} else if now.Sub(lastRegisterTime) > 200*time.Millisecond {
						v2["health"] = false
					}
				}
			}
			timer.Reset(interval)
		}
	}
}

func (cn *ClusterNode) RPCRegister(args *Node, reply *Node) {
	slog.Info(fmt.Sprintf("Register from Cluster Node: `%s:%s`", args.Id, args.Endpoint))
	slog.Info(fmt.Sprintf("Cluster Node Scheduler RPC Service listening at: %s", args.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Node Scheduler HTTP Service listening at: %s", args.EndpointHTTP))
	slog.Info(fmt.Sprintf("Cluster Node Queue: `%s`", args.Queue))

	cn.registerNode(args.toClusterNode())

	reply.Id = cn.Id
	reply.MainEndpoint = cn.MainEndpoint
	reply.Endpoint = cn.Endpoint
	reply.EndpointHTTP = cn.EndpointHTTP
	reply.SchedulerEndpoint = cn.SchedulerEndpoint
	reply.Queue = cn.Queue

	reply.QueueMap = cn.queueMap
}

func (cn *ClusterNode) RPCPing(args *Node, reply *Node) {
	cn.registerNode(args.toClusterNode())

	reply.QueueMap = cn.queueMap
}

func (cn *ClusterNode) RegisterNodeRemote(ctx context.Context) error {
	slog.Info(fmt.Sprintf("Register to Cluster Main Node: `%s`", cn.MainEndpoint))

	rClient, err := rpc.DialHTTP("tcp", cn.MainEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster main node: `%s`, error: %s", cn.MainEndpoint, err)
	}

	var main Node
	ch := make(chan error, 1)
	go func() { ch <- rClient.Call("CRPCService.Register", cn.toNode(), &main) }()
	select {
	case err := <-ch:
		if err != nil {
			return fmt.Errorf("failed to register to cluster main node, error: %s", err)
		}
	case <-time.After(3 * time.Second):
		return fmt.Errorf("register to cluster main node `%s` timeout", cn.MainEndpoint)
	}
	cn.setQueueMap(main.QueueMap)

	slog.Info(fmt.Sprintf("Cluster Main Node Scheduler RPC Service listening at: %s", main.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Main Node Scheduler http Service listening at: %s", main.EndpointHTTP))
	slog.Info(fmt.Sprintf("Cluster Main Node Queue: `%s`", main.Queue))

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
			if err := cn.pingRemote(ctx); err != nil {
				slog.Info(fmt.Sprintf("Ping remote error: %s", err))
				timer.Reset(time.Second)
			} else {
				timer.Reset(interval)
			}
		}
	}
}

func (cn *ClusterNode) pingRemote(ctx context.Context) error {
	rClient, err := rpc.DialHTTP("tcp", cn.MainEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster main node: `%s`, error: %s", cn.MainEndpoint, err)
	}

	var main Node
	ch := make(chan error, 1)
	go func() { ch <- rClient.Call("CRPCService.Ping", cn.toNode(), &main) }()
	select {
	case err := <-ch:
		if err != nil {
			return fmt.Errorf("failed to ping to cluster main node, error: %s", err)
		}
	case <-time.After(200 * time.Millisecond):
		return fmt.Errorf("ping to cluster main node `%s` timeout", cn.MainEndpoint)
	}
	cn.setQueueMap(main.QueueMap)

	return nil
}
