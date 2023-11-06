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

// Each node provides `RPC`, `HTTP`, `Scheduler gRPC` services,
// but only the main node starts the scheduler,
// the other worker nodes register with the main node
// and then run jobs from the main node via the Scheduler's `RunJob` API.
type ClusterNode struct {
	// The unique identifier of this node, automatically generated.
	// It should not be set manually.
	Id string
	// Main node RPC listening address.
	// If you are the main, `MainEndpoint` is the same as `Endpoint`.
	// Default: `127.0.0.1:36364`
	MainEndpoint string
	// RPC listening address.
	// Used to expose the cluster's internal API.
	// Default: `127.0.0.1:36364`
	Endpoint string
	// HTTP listening address.
	// Used to expose the cluster's external API.
	// Default: `127.0.0.1:63637`
	EndpointHTTP string
	// Scheduler gRPC listening address.
	// Used to expose the scheduler's external API.
	// Default: `127.0.0.1:36363`
	SchedulerEndpoint string
	// Useful when a job specifies a queue.
	// A queue number can correspond to multiple nodes.
	// Default: `default`
	Queue string

	// Stores node information for the entire cluster.
	// It should not be set manually.
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

// Initialization functions for each node,
// called when the scheduler run `SetClusterNode`.
func (cn *ClusterNode) init(ctx context.Context) error {
	cn.setId()
	cn.registerNode(cn)

	if cn.MainEndpoint == cn.Endpoint {
		go cn.checkNode(ctx)
	}

	return nil
}

// Register node with the cluster.
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

// Randomly select a healthy node from the cluster,
// if you specify a queue number, filter by queue number.
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

// Regularly check node,
// if a node has not been updated for a long time it is marked as unhealthy or the node is deleted.
func (cn *ClusterNode) checkNode(ctx context.Context) {
	interval := 400 * time.Millisecond
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
					} else if now.Sub(lastRegisterTime) > 400*time.Millisecond {
						v2["health"] = false
					}
				}
			}
			timer.Reset(interval)
		}
	}
}

// RPC API
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

// RPC API
func (cn *ClusterNode) RPCPing(args *Node, reply *Node) {
	cn.registerNode(args.toClusterNode())

	reply.QueueMap = cn.queueMap
}

// Used for work node
//
// After initialization, node need to register with the main node and synchronize cluster node information.
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
	slog.Info(fmt.Sprintf("Cluster Main Node Scheduler HTTP Service listening at: %s", main.EndpointHTTP))
	slog.Info(fmt.Sprintf("Cluster Main Node Queue: `%s`", main.Queue))

	go cn.heartbeatRemote(ctx)

	return nil
}

// Used for work node
//
// Started when the node run `RegisterNodeRemote`.
func (cn *ClusterNode) heartbeatRemote(ctx context.Context) {
	interval := 200 * time.Millisecond
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

// Used for work node
//
// Update and synchronize cluster node information.
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
	case <-time.After(400 * time.Millisecond):
		return fmt.Errorf("ping to cluster main node `%s` timeout", cn.MainEndpoint)
	}
	cn.setQueueMap(main.QueueMap)

	return nil
}
