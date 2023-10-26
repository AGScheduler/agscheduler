package agscheduler

import (
	"fmt"
	"log/slog"
	"net/rpc"
	"strings"

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
	cn.registerMain()

	return nil
}

func (cn *ClusterNode) Register(args *Node, reply *Node) error {
	slog.Info(fmt.Sprintf("Registration from the cluster node `%s:%s`:", args.Id, args.Endpoint))
	slog.Info(fmt.Sprintf("Cluster Node Scheduler RPC Service listening at: %s", args.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Node Scheduler RPC Service queue: `%s`", args.SchedulerQueue))

	if _, ok := cn.queueMap[args.SchedulerQueue]; !ok {
		cn.queueMap[args.SchedulerQueue] = map[string]map[string]any{}
	}
	cn.queueMap[args.SchedulerQueue][args.Id] = map[string]any{
		"endpoint":           args.Endpoint,
		"scheduler_endpoint": args.SchedulerEndpoint,
		"health":             true,
	}

	reply.Id = cn.Id
	reply.Endpoint = cn.Endpoint
	reply.SchedulerEndpoint = cn.SchedulerEndpoint
	reply.SchedulerQueue = cn.SchedulerQueue
	reply.queueMap = cn.queueMap

	return nil
}

func (cn *ClusterNode) registerMain() error {
	cn.queueMap[cn.SchedulerQueue] = map[string]map[string]any{}
	cn.queueMap[cn.SchedulerQueue][cn.Id] = map[string]any{
		"endpoint":           cn.Endpoint,
		"scheduler_endpoint": cn.SchedulerEndpoint,
		"health":             true,
	}

	return nil
}

func (cn *ClusterNode) RegisterNode() error {
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
	err = rClient.Call("CRPCService.Register", node, &main)
	if err != nil {
		return fmt.Errorf("failed to register to cluster main, error: %s", err)
	}
	cn.queueMap = main.queueMap

	slog.Info(fmt.Sprintf("Cluster Main Scheduler RPC Service listening at: %s", main.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Main Scheduler RPC Service queue: `%s`", main.SchedulerQueue))

	return nil
}
