package agscheduler

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"
	"math/rand"
	"net/rpc"
	"slices"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/agscheduler/agscheduler/services/proto"
)

type TypeNodeMap map[string]map[string]any

var nodeMapMutexC sync.RWMutex
var endpointMainMutexC sync.RWMutex

type Node struct {
	EndpointMain string
	Endpoint     string
	EndpointGRPC string
	EndpointHTTP string
	Queue        string
	Mode         string

	NodeMap TypeNodeMap
}

func (n *Node) toClusterNode() *ClusterNode {
	return &ClusterNode{
		EndpointMain: n.EndpointMain,
		Endpoint:     n.Endpoint,
		EndpointGRPC: n.EndpointGRPC,
		EndpointHTTP: n.EndpointHTTP,
		Queue:        n.Queue,
		Mode:         n.Mode,

		nodeMap: n.NodeMap,
	}
}

// Each node provides `Cluster RPC`, `gRPC`, `HTTP` services,
// but only the main node starts the scheduler,
// the other worker nodes register with the main node
// and then run jobs from the main node via the RPC's `RunJob` API.
type ClusterNode struct {
	// Main node RPC listening address.
	// If you are the main, `EndpointMain` is the same as `Endpoint`.
	// Default: `127.0.0.1:36380`
	EndpointMain string
	// The unique identifier of this node.
	// RPC listening address.
	// Used to expose the cluster's internal API.
	// Default: `127.0.0.1:36380`
	Endpoint string
	// gRPC listening address.
	// Used to expose the external API.
	// Default: `127.0.0.1:36360`
	EndpointGRPC string
	// HTTP listening address.
	// Used to expose the external API.
	// Default: `127.0.0.1:36370`
	EndpointHTTP string
	// Useful when a job specifies a queue.
	// A queue can correspond to multiple nodes.
	// Default: `default`
	Queue string
	// Node mode, for Scheduler high availability.
	// If the value is `HA`, the node will join the raft group.
	// Default: ``, Options `HA`
	Mode string

	// Stores node information for the entire cluster.
	// It should not be set manually.
	// def: map[<endpoint>]map[string]any
	nodeMap TypeNodeMap

	// Bind to each other and the Scheduler.
	Scheduler *Scheduler

	// For Scheduler high availability.
	// Bind to each other and the Raft.
	Raft *Raft
	// Used to mark the status of Cluster Scheduler operation.
	SchedulerCanStart bool
}

func (cn *ClusterNode) toNode() *Node {
	return &Node{
		EndpointMain: cn.GetEndpointMain(),
		Endpoint:     cn.Endpoint,
		EndpointGRPC: cn.EndpointGRPC,
		EndpointHTTP: cn.EndpointHTTP,
		Queue:        cn.Queue,
		Mode:         cn.Mode,

		NodeMap: cn.NodeMapCopy(),
	}
}

func (cn *ClusterNode) setNodeMap(nmap TypeNodeMap) {
	nodeMapMutexC.Lock()
	defer nodeMapMutexC.Unlock()

	cn.nodeMap = nmap
}

func (cn *ClusterNode) deepCopyNodeMapByGob(dst, src TypeNodeMap) error {
	gob.Register(time.Time{})

	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(src); err != nil {
		return err
	}

	return gob.NewDecoder(bytes.NewBuffer(buffer.Bytes())).Decode(&dst)
}

func (cn *ClusterNode) NodeMapCopy() TypeNodeMap {
	nodeMapMutexC.RLock()
	defer nodeMapMutexC.RUnlock()

	nodeMapCopy := make(TypeNodeMap)
	err := cn.deepCopyNodeMapByGob(nodeMapCopy, cn.nodeMap)
	if err != nil {
		slog.Error("Deep copy `NodeMap` error:", err)
	}

	return nodeMapCopy
}

// Used to gRPC Protobuf
func (cn *ClusterNode) NodeMapToPbNodesPtr() *pb.Nodes {
	pbN := pb.Nodes{}
	pbN.Nodes = make(map[string]*pb.Node)

	for k, v := range cn.NodeMapCopy() {
		pbN.Nodes[k] = &pb.Node{
			EndpointMain:      v["endpoint_main"].(string),
			Endpoint:          v["endpoint"].(string),
			EndpointGrpc:      v["endpoint_grpc"].(string),
			EndpointHttp:      v["endpoint_http"].(string),
			Queue:             v["queue"].(string),
			Mode:              v["mode"].(string),
			Version:           Version,
			Health:            v["health"].(bool),
			RegisterTime:      timestamppb.New(v["register_time"].(time.Time)),
			LastHeartbeatTime: timestamppb.New(v["last_heartbeat_time"].(time.Time)),
		}
	}

	return &pbN
}

func (cn *ClusterNode) MainNode() map[string]any {
	return cn.NodeMapCopy()[cn.GetEndpointMain()]
}

func (cn *ClusterNode) HANodeMap() TypeNodeMap {
	HANodeMap := make(TypeNodeMap)
	for endpoint, v := range cn.NodeMapCopy() {
		if v["mode"].(string) != "HA" {
			continue
		}
		HANodeMap[endpoint] = v
	}

	return HANodeMap
}

func (cn *ClusterNode) SetEndpointMain(endpoint string) {
	endpointMainMutexC.Lock()
	defer endpointMainMutexC.Unlock()

	cn.EndpointMain = endpoint
}

func (cn *ClusterNode) GetEndpointMain() string {
	endpointMainMutexC.RLock()
	defer endpointMainMutexC.RUnlock()

	return cn.EndpointMain
}

func (cn *ClusterNode) IsMainNode() bool {
	return cn.GetEndpointMain() == cn.Endpoint
}

// Initialization functions for each node,
// called when the scheduler run `SetClusterNode`.
func (cn *ClusterNode) init(ctx context.Context) error {
	if cn.Endpoint == "" {
		cn.Endpoint = "127.0.0.1:36380"
	}
	if cn.GetEndpointMain() == "" {
		cn.SetEndpointMain(cn.Endpoint)
	}
	if cn.EndpointGRPC == "" {
		cn.EndpointGRPC = "127.0.0.1:36360"
	}
	if cn.EndpointHTTP == "" {
		cn.EndpointHTTP = "127.0.0.1:36370"
	}
	if cn.Queue == "" {
		cn.Queue = "default"
	}
	cn.Mode = strings.ToUpper(cn.Mode)

	cn.registerNode(cn)

	go cn.heartbeat(ctx)
	go cn.checkNode(ctx)

	if cn.Mode == "HA" {
		cn.Raft = &Raft{cn: cn}
		cn.Raft.start(ctx)
	}

	return nil
}

// Register node with the cluster.
func (cn *ClusterNode) registerNode(n *ClusterNode) {
	nodeMapMutexC.Lock()
	defer nodeMapMutexC.Unlock()

	if cn.nodeMap == nil {
		cn.nodeMap = make(TypeNodeMap)
	}
	if _, ok := cn.nodeMap[n.Endpoint]; !ok {
		cn.nodeMap[n.Endpoint] = map[string]any{}
	}
	now := time.Now().UTC()
	register_time := cn.nodeMap[n.Endpoint]["register_time"]
	if register_time == nil {
		register_time = now
	}
	cn.nodeMap[n.Endpoint] = map[string]any{
		"endpoint_main":       n.GetEndpointMain(),
		"endpoint":            n.Endpoint,
		"endpoint_grpc":       n.EndpointGRPC,
		"endpoint_http":       n.EndpointHTTP,
		"queue":               n.Queue,
		"mode":                n.Mode,
		"version":             Version,
		"health":              true,
		"register_time":       register_time,
		"last_heartbeat_time": now,
	}
}

// Randomly select a healthy node from the cluster,
// if you specify a queue, filter by queue.
func (cn *ClusterNode) choiceNode(queues []string) (*ClusterNode, error) {
	cns := make([]*ClusterNode, 0)
	for endpoint, v := range cn.NodeMapCopy() {
		if !v["health"].(bool) {
			continue
		}
		if len(queues) != 0 && !slices.Contains(queues, v["queue"].(string)) {
			continue
		}
		cns = append(cns, &ClusterNode{
			EndpointMain: v["endpoint_main"].(string),
			Endpoint:     endpoint,
			EndpointGRPC: v["endpoint_grpc"].(string),
			EndpointHTTP: v["endpoint_http"].(string),
			Queue:        v["queue"].(string),
			Mode:         v["mode"].(string),
		})
	}

	cnsCount := len(cns)
	if cnsCount != 0 {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		i := rand.Intn(cnsCount)
		return cns[i], nil
	}

	return &ClusterNode{}, fmt.Errorf("cluster node not found")
}

// Maintaining node health.
// Started when the node init.
func (cn *ClusterNode) heartbeat(ctx context.Context) {
	interval := 400 * time.Millisecond
	timer := time.NewTimer(interval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if err := cn.heartbeatRemote(ctx); err != nil {
				slog.Info(fmt.Sprintf("Send heartbeat remote error: %s", err))
				timer.Reset(time.Second)
			} else {
				timer.Reset(interval)
			}
		}
	}
}

// Regularly check node,
// if a node has not been updated for a long time it is marked as unhealthy or the node is deleted.
// HA nodes are not processed.
// Started when the node init.
func (cn *ClusterNode) checkNode(ctx context.Context) {
	interval := 600 * time.Millisecond
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
			for endpoint, v := range cn.NodeMapCopy() {
				if cn.Endpoint == endpoint {
					continue
				}
				lastHeartbeatTime := v["last_heartbeat_time"].(time.Time)
				if now.Sub(lastHeartbeatTime) > 5*time.Minute {
					if v["mode"].(string) == "HA" {
						continue
					}
					nodeMapMutexC.Lock()
					delete(cn.nodeMap, endpoint)
					nodeMapMutexC.Unlock()
					slog.Warn(fmt.Sprintf("Cluster node `%s` have been deleted because unhealthy", endpoint))
				} else if now.Sub(lastHeartbeatTime) > 600*time.Millisecond {
					nodeMapMutexC.Lock()
					cn.nodeMap[endpoint]["health"] = false
					nodeMapMutexC.Unlock()
				}
			}
			timer.Reset(interval)
		}
	}
}

// RPC API
func (cn *ClusterNode) RPCRegister(args *Node, reply *Node) {
	slog.Info(fmt.Sprintf("Register from Cluster Node: `%s`", args.Endpoint))
	slog.Info(fmt.Sprintf("Cluster Node gRPC Service listening at: %s", args.EndpointGRPC))
	slog.Info(fmt.Sprintf("Cluster Node HTTP Service listening at: %s", args.EndpointHTTP))
	slog.Info(fmt.Sprintf("Cluster Node Queue: `%s`", args.Queue))

	cn.registerNode(args.toClusterNode())

	reply.EndpointMain = cn.GetEndpointMain()
	reply.Endpoint = cn.Endpoint
	reply.EndpointGRPC = cn.EndpointGRPC
	reply.EndpointHTTP = cn.EndpointHTTP
	reply.Queue = cn.Queue

	reply.NodeMap = cn.NodeMapCopy()
}

// RPC API
func (cn *ClusterNode) RPCHeartbeat(args *Node, reply *Node) {
	cn.registerNode(args.toClusterNode())

	reply.EndpointMain = cn.GetEndpointMain()
	reply.NodeMap = cn.NodeMapCopy()
}

// Used for worker node
//
// After initialization, node need to register with the main node and synchronize cluster node information.
func (cn *ClusterNode) RegisterNodeRemote(ctx context.Context) error {
	slog.Info(fmt.Sprintf("Register to Cluster Main Node: `%s`", cn.GetEndpointMain()))

	rClient, err := rpc.DialHTTP("tcp", cn.GetEndpointMain())
	if err != nil {
		return fmt.Errorf("failed to connect to cluster main node: `%s`, error: %s", cn.GetEndpointMain(), err)
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
		return fmt.Errorf("register to cluster main node `%s` timeout", cn.GetEndpointMain())
	}
	cn.SetEndpointMain(main.EndpointMain)
	cn.setNodeMap(main.NodeMap)

	slog.Info(fmt.Sprintf("Cluster Main Node gRPC Service listening at: %s", main.EndpointGRPC))
	slog.Info(fmt.Sprintf("Cluster Main Node HTTP Service listening at: %s", main.EndpointHTTP))
	slog.Info(fmt.Sprintf("Cluster Main Node Queue: `%s`", main.Queue))

	return nil
}

// Used for worker node
//
// Update and synchronize cluster node information.
func (cn *ClusterNode) heartbeatRemote(ctx context.Context) error {
	rClient, err := rpc.DialHTTP("tcp", cn.GetEndpointMain())
	if err != nil {
		return fmt.Errorf("failed to connect to cluster main node: `%s`, error: %s", cn.GetEndpointMain(), err)
	}
	defer rClient.Close()

	var main Node
	ch := make(chan error, 1)
	go func() { ch <- rClient.Call("CRPCService.Heartbeat", cn.toNode(), &main) }()
	select {
	case err := <-ch:
		if err != nil {
			return fmt.Errorf("failed to send heartbeat to cluster main node, error: %s", err)
		}
	case <-time.After(300 * time.Millisecond):
		return fmt.Errorf("send heartbeat to cluster main node `%s` timeout", cn.GetEndpointMain())
	}
	cn.SetEndpointMain(main.EndpointMain)
	cn.setNodeMap(main.NodeMap)

	return nil
}
