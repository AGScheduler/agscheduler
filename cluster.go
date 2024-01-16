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
)

var mutexC sync.RWMutex

type Node struct {
	MainEndpoint          string
	Endpoint              string
	EndpointHTTP          string
	SchedulerEndpoint     string
	SchedulerEndpointHTTP string
	Queue                 string
	Mode                  string
	NodeMap               map[string]map[string]map[string]any
}

func (n *Node) toClusterNode() *ClusterNode {
	return &ClusterNode{
		MainEndpoint:          n.MainEndpoint,
		Endpoint:              n.Endpoint,
		EndpointHTTP:          n.EndpointHTTP,
		SchedulerEndpoint:     n.SchedulerEndpoint,
		SchedulerEndpointHTTP: n.SchedulerEndpointHTTP,
		Queue:                 n.Queue,
		Mode:                  n.Mode,

		nodeMap: n.NodeMap,
	}
}

// Each node provides `RPC`, `HTTP`, `Scheduler gRPC` services,
// but only the main node starts the scheduler,
// the other worker nodes register with the main node
// and then run jobs from the main node via the Scheduler's `RunJob` API.
type ClusterNode struct {
	// Main node RPC listening address.
	// If you are the main, `MainEndpoint` is the same as `Endpoint`.
	// Default: `127.0.0.1:36380`
	MainEndpoint string
	// The unique identifier of this node.
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
	// Scheduler HTTP listening address.
	// Used to expose the scheduler's external API.
	// Default: `127.0.0.1:36370`
	SchedulerEndpointHTTP string
	// Useful when a job specifies a queue.
	// A queue can correspond to multiple nodes.
	// Default: `default`
	Queue string

	Mode string

	// Stores node information for the entire cluster.
	// It should not be set manually.
	// def: map[<queue>]map[<endpoint>]map[string]any
	nodeMap map[string]map[string]map[string]any

	// Bind to each other and the scheduler.
	Scheduler *Scheduler

	Raft *Raft
}

func (cn *ClusterNode) toNode() *Node {
	return &Node{
		MainEndpoint:          cn.MainEndpoint,
		Endpoint:              cn.Endpoint,
		EndpointHTTP:          cn.EndpointHTTP,
		SchedulerEndpoint:     cn.SchedulerEndpoint,
		SchedulerEndpointHTTP: cn.SchedulerEndpointHTTP,
		Queue:                 cn.Queue,
		Mode:                  cn.Mode,
		NodeMap:               cn.NodeMap(),
	}
}

func (cn *ClusterNode) setNodeMap(nmap map[string]map[string]map[string]any) {
	defer mutexC.Unlock()

	mutexC.Lock()
	cn.nodeMap = nmap
}

func (cn *ClusterNode) NodeMap() map[string]map[string]map[string]any {
	defer mutexC.RUnlock()

	mutexC.RLock()
	return cn.nodeMap
}

func (cn *ClusterNode) MainNode() map[string]any {
	for _, v := range cn.NodeMap() {
		for endpoint, v2 := range v {
			if cn.MainEndpoint != endpoint {
				continue
			}
			return v2
		}
	}

	return make(map[string]any)
}

func (cn *ClusterNode) HANodeMap() map[string]map[string]map[string]any {
	HANodeMap := make(map[string]map[string]map[string]any)
	for queue, v := range cn.NodeMap() {
		for endpoint, v2 := range v {
			if v2["mode"].(string) != "HA" {
				continue
			}
			if _, ok := HANodeMap[queue]; !ok {
				HANodeMap[queue] = map[string]map[string]any{}
			}
			HANodeMap[queue][endpoint] = v2
		}
	}

	return HANodeMap
}

func (cn *ClusterNode) IsMainNode() bool {
	return cn.MainEndpoint == cn.Endpoint
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
	if cn.SchedulerEndpointHTTP == "" {
		cn.SchedulerEndpointHTTP = "127.0.0.1:36370"
	}
	if cn.Queue == "" {
		cn.Queue = "default"
	}
	cn.Mode = strings.ToUpper(cn.Mode)

	cn.registerNode(cn)

	go cn.heartbeatRemote(ctx)
	go cn.checkNode(ctx)

	if cn.Mode == "HA" {
		cn.Raft = &Raft{cn: cn}
		cn.Raft.start(ctx)
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
	register_time := cn.nodeMap[n.Queue][n.Endpoint]["register_time"]
	if register_time == nil {
		register_time = now
	}
	cn.nodeMap[n.Queue][n.Endpoint] = map[string]any{
		"main_endpoint":           n.MainEndpoint,
		"endpoint":                n.Endpoint,
		"endpoint_http":           n.EndpointHTTP,
		"scheduler_endpoint":      n.SchedulerEndpoint,
		"scheduler_endpoint_http": n.SchedulerEndpointHTTP,
		"queue":                   n.Queue,
		"mode":                    n.Mode,
		"health":                  true,
		"register_time":           register_time,
		"last_heartbeat_time":     now,
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
		for endpoint, v2 := range v {
			if !v2["health"].(bool) {
				continue
			}
			cns = append(cns, &ClusterNode{
				MainEndpoint:          v2["main_endpoint"].(string),
				Endpoint:              endpoint,
				EndpointHTTP:          v2["endpoint_http"].(string),
				SchedulerEndpoint:     v2["scheduler_endpoint"].(string),
				SchedulerEndpointHTTP: v2["scheduler_endpoint_http"].(string),
				Queue:                 v2["queue"].(string),
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
			if !cn.IsMainNode() {
				timer.Reset(interval)
				continue
			}

			now := time.Now().UTC()
			for _, v := range cn.NodeMap() {
				for endpoint, v2 := range v {
					if cn.Endpoint == endpoint {
						continue
					}
					lastHeartbeatTime := v2["last_heartbeat_time"].(time.Time)
					if now.Sub(lastHeartbeatTime) > 5*time.Minute {
						mutexC.Lock()
						delete(v, endpoint)
						mutexC.Unlock()
						slog.Warn(fmt.Sprintf("Cluster node `%s` have been deleted because unhealthy", endpoint))
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
	slog.Info(fmt.Sprintf("Register from Cluster Node: `%s`", args.Endpoint))
	slog.Info(fmt.Sprintf("Cluster Node HTTP Service listening at: %s", args.EndpointHTTP))
	slog.Info(fmt.Sprintf("Cluster Node Scheduler gRPC Service listening at: %s", args.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Node Scheduler HTTP Service listening at: %s", args.SchedulerEndpointHTTP))
	slog.Info(fmt.Sprintf("Cluster Node Queue: `%s`", args.Queue))

	cn.registerNode(args.toClusterNode())

	reply.MainEndpoint = cn.MainEndpoint
	reply.Endpoint = cn.Endpoint
	reply.EndpointHTTP = cn.EndpointHTTP
	reply.SchedulerEndpoint = cn.SchedulerEndpoint
	reply.SchedulerEndpointHTTP = cn.SchedulerEndpointHTTP
	reply.Queue = cn.Queue

	reply.NodeMap = cn.NodeMap()
}

// RPC API
func (cn *ClusterNode) RPCPing(args *Node, reply *Node) {
	cn.registerNode(args.toClusterNode())

	reply.MainEndpoint = cn.MainEndpoint
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
	cn.MainEndpoint = main.MainEndpoint
	cn.setNodeMap(main.NodeMap)

	slog.Info(fmt.Sprintf("Cluster Main Node HTTP Service listening at: %s", main.EndpointHTTP))
	slog.Info(fmt.Sprintf("Cluster Main Node Scheduler gRPC Service listening at: %s", main.SchedulerEndpoint))
	slog.Info(fmt.Sprintf("Cluster Main Node Scheduler HTTP Service listening at: %s", main.SchedulerEndpointHTTP))
	slog.Info(fmt.Sprintf("Cluster Main Node Queue: `%s`", main.Queue))

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
			if cn.IsMainNode() {
				timer.Reset(interval)
				continue
			}

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
	cn.MainEndpoint = main.MainEndpoint
	cn.setNodeMap(main.NodeMap)

	return nil
}
