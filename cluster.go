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

var mutexC sync.Mutex

type Node struct {
	Id                string
	MainEndpoint      string
	Endpoint          string
	EndpointHTTP      string
	SchedulerEndpoint string
	Queue             string
	NodeMap           map[string]map[string]map[string]any
}

func (n *Node) toClusterNode() *ClusterNode {
	return &ClusterNode{
		Id:                n.Id,
		MainEndpoint:      n.MainEndpoint,
		Endpoint:          n.Endpoint,
		EndpointHTTP:      n.EndpointHTTP,
		SchedulerEndpoint: n.SchedulerEndpoint,
		Queue:             n.Queue,

		nodeMap: n.NodeMap,
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
	// Default: `127.0.0.1:36380`
	MainEndpoint string
	// RPC listening address.
	// Used to expose the cluster's internal API.
	// Default: `127.0.0.1:36380`
	Endpoint string
	// HTTP listening address.
	// Used to expose the cluster's external API.
	// Default: `127.0.0.1:36390`
	EndpointHTTP string
	// Scheduler gRPC listening address.
	// Used to expose the scheduler's external API.
	// Default: `127.0.0.1:36360`
	SchedulerEndpoint string
	// Useful when a job specifies a queue.
	// A queue can correspond to multiple nodes.
	// Default: `default`
	Queue string

	// Stores node information for the entire cluster.
	// It should not be set manually.
	// def: map[<queue>]map[<id>]map[string]any
	nodeMap map[string]map[string]map[string]any

	// Bind to each other and the scheduler.
	Scheduler *Scheduler
}

func (cn *ClusterNode) toNode() *Node {
	return &Node{
		Id:                cn.Id,
		MainEndpoint:      cn.MainEndpoint,
		Endpoint:          cn.Endpoint,
		EndpointHTTP:      cn.EndpointHTTP,
		SchedulerEndpoint: cn.SchedulerEndpoint,
		Queue:             cn.Queue,
		NodeMap:           cn.NodeMap(),
	}
}

func (cn *ClusterNode) setNodeMap(nmap map[string]map[string]map[string]any) {
	defer mutexC.Unlock()

	mutexC.Lock()
	cn.nodeMap = nmap
}

func (cn *ClusterNode) NodeMap() map[string]map[string]map[string]any {
	defer mutexC.Unlock()

	mutexC.Lock()
	return cn.nodeMap
}

func (cn *ClusterNode) setId() {
	cn.Id = strings.Replace(uuid.New().String(), "-", "", -1)[:16]
}

// Initialization functions for each node,
// called when the scheduler run `SetClusterNode`.
func (cn *ClusterNode) init(ctx context.Context) error {
	if cn.Endpoint == "" {
		cn.Endpoint = "127.0.0.1:36380"
	}
	if cn.MainEndpoint == "" {
		cn.MainEndpoint = cn.Endpoint
	}
	if cn.EndpointHTTP == "" {
		cn.EndpointHTTP = "127.0.0.1:36390"
	}
	if cn.SchedulerEndpoint == "" {
		cn.SchedulerEndpoint = "127.0.0.1:36360"
	}
	if cn.Queue == "" {
		cn.Queue = "default"
	}

	cn.setId()
	cn.registerNode(cn)

	if cn.MainEndpoint == cn.Endpoint {
		go cn.checkNode(ctx)
	}

	return nil
}

// Register node with the cluster.
func (cn *ClusterNode) registerNode(n *ClusterNode) {
	defer mutexC.Unlock()

	mutexC.Lock()

	if cn.nodeMap == nil {
		cn.nodeMap = make(map[string]map[string]map[string]any)
	}
	if _, ok := cn.nodeMap[n.Queue]; !ok {
		cn.nodeMap[n.Queue] = map[string]map[string]any{}
	}
	now := time.Now().UTC()
	register_time := cn.nodeMap[n.Queue][n.Id]["register_time"]
	if register_time == nil {
		register_time = now
	}
	cn.nodeMap[n.Queue][n.Id] = map[string]any{
		"id":                  n.Id,
		"main_endpoint":       n.MainEndpoint,
		"endpoint":            n.Endpoint,
		"endpoint_http":       n.EndpointHTTP,
		"scheduler_endpoint":  n.SchedulerEndpoint,
		"queue":               n.Queue,
		"health":              true,
		"register_time":       register_time,
		"last_heartbeat_time": now,
	}
}

// Randomly select a healthy node from the cluster,
// if you specify a queue, filter by queue.
func (cn *ClusterNode) choiceNode(queues []string) (*ClusterNode, error) {
	cns := make([]*ClusterNode, 0)
	for q, v := range cn.NodeMap() {
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
			for _, v := range cn.NodeMap() {
				for id, v2 := range v {
					if cn.Id == id {
						continue
					}
					endpoint := v2["endpoint"].(string)
					lastHeartbeatTime := v2["last_heartbeat_time"].(time.Time)
					if now.Sub(lastHeartbeatTime) > 5*time.Minute {
						mutexC.Lock()
						delete(v, id)
						mutexC.Unlock()
						slog.Warn(fmt.Sprintf("Cluster node `%s:%s` have been deleted because unhealthy", id, endpoint))
					} else if now.Sub(lastHeartbeatTime) > 400*time.Millisecond {
						mutexC.Lock()
						v2["health"] = false
						mutexC.Unlock()
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

	reply.NodeMap = cn.NodeMap()
}

// RPC API
func (cn *ClusterNode) RPCPing(args *Node, reply *Node) {
	cn.registerNode(args.toClusterNode())

	reply.NodeMap = cn.NodeMap()
}

// Used for worker node
//
// After initialization, node need to register with the main node and synchronize cluster node information.
func (cn *ClusterNode) RegisterNodeRemote(ctx context.Context) error {
	slog.Info(fmt.Sprintf("Register to Cluster Main Node: `%s`", cn.MainEndpoint))

	rClient, err := rpc.DialHTTP("tcp", cn.MainEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster main node: `%s`, error: %s", cn.MainEndpoint, err)
	}
	defer rClient.Close()

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
	cn.setNodeMap(main.NodeMap)

	slog.Info(fmt.Sprintf("Cluster Main Node Scheduler RPC Service listening at: %s", main.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Main Node Scheduler HTTP Service listening at: %s", main.EndpointHTTP))
	slog.Info(fmt.Sprintf("Cluster Main Node Queue: `%s`", main.Queue))

	go cn.heartbeatRemote(ctx)

	return nil
}

// Used for worker node
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

// Used for worker node
//
// Update and synchronize cluster node information.
func (cn *ClusterNode) pingRemote(ctx context.Context) error {
	rClient, err := rpc.DialHTTP("tcp", cn.MainEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster main node: `%s`, error: %s", cn.MainEndpoint, err)
	}
	defer rClient.Close()

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
	cn.setNodeMap(main.NodeMap)

	return nil
}
