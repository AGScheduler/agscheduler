package agscheduler

import (
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

type ClusterNode struct {
	Id                string
	MainEndpoint      string
	Endpoint          string
	SchedulerEndpoint string
	SchedulerQueue    string
	queueMap          map[string]map[string]map[string]any
}

func (cn *ClusterNode) setId() {
	cn.Id = strings.Replace(uuid.New().String(), "-", "", -1)[:16]
}

func (cn *ClusterNode) init() error {
	cn.setId()
	cn.queueMap = make(map[string]map[string]map[string]any)
	cn.registerNode(cn)

	return nil
}

func (cn *ClusterNode) RPCRegister(args *Node, reply *Node) error {
	slog.Info(fmt.Sprintf("Registration from the cluster node `%s:%s`:", args.Id, args.Endpoint))
	slog.Info(fmt.Sprintf("Cluster Node Scheduler RPC Service listening at: %s", args.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Node Scheduler RPC Service queue: `%s`", args.SchedulerQueue))

	cn.registerNode(&ClusterNode{
		Id:                args.Id,
		MainEndpoint:      args.MainEndpoint,
		Endpoint:          args.Endpoint,
		SchedulerEndpoint: args.SchedulerEndpoint,
		SchedulerQueue:    args.SchedulerQueue,
	})

	reply.Id = cn.Id
	reply.Endpoint = cn.Endpoint
	reply.SchedulerEndpoint = cn.SchedulerEndpoint
	reply.SchedulerQueue = cn.SchedulerQueue
	reply.queueMap = cn.queueMap

	return nil
}

func (cn *ClusterNode) registerNode(n *ClusterNode) error {
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
	}

	return nil
}

func (cn *ClusterNode) RegisterNodeRemote() error {
	slog.Info(fmt.Sprintf("Register with cluster main `%s`:", cn.MainEndpoint))

	rClient, err := rpc.DialHTTP("tcp", cn.MainEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster main: `%s`, error: %s", cn.MainEndpoint, err)
	}

	node := Node{
		Id:                cn.Id,
		MainEndpoint:      cn.MainEndpoint,
		Endpoint:          cn.Endpoint,
		SchedulerEndpoint: cn.SchedulerEndpoint,
		SchedulerQueue:    cn.SchedulerQueue,
	}
	var main Node
	c := make(chan error, 1)
	go func() { c <- rClient.Call("CRPCService.Register", node, &main) }()
	select {
	case err := <-c:
		if err != nil {
			return fmt.Errorf("failed to register to cluster main, error: %s", err)
		}
	case <-time.After(3 * time.Second):
		return fmt.Errorf("register to cluster main timeout: %s", err)
	}

	cn.queueMap = main.queueMap

	slog.Info(fmt.Sprintf("Cluster Main Scheduler RPC Service listening at: %s", main.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Main Scheduler RPC Service queue: `%s`", main.SchedulerQueue))

	return nil
}

func (cn *ClusterNode) choiceNode() (*ClusterNode, error) {
	cns := make([]*ClusterNode, 0)
	for _, v := range cn.queueMap {
		for _, v2 := range v {
			if !v2["health"].(bool) {
				continue
			}
			cns = append(cns, &ClusterNode{
				Id:                v2["id"].(string),
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

	return &ClusterNode{}, fmt.Errorf("node not found")
}
