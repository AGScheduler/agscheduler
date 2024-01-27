package agscheduler

import (
	"context"
	"encoding/gob"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	pb "github.com/kwkwc/agscheduler/services/proto"
)

func getClusterNode() *ClusterNode {
	return &ClusterNode{
		EndpointMain: "127.0.0.1:36380",
		Endpoint:     "127.0.0.1:36380",
		EndpointGRPC: "127.0.0.1:36360",
		EndpointHTTP: "127.0.0.1:36370",
		Queue:        "default",
		Mode:         "",
	}
}

func TestClusterToClusterNode(t *testing.T) {
	n := Node{
		EndpointMain: "127.0.0.1:36380",
		Endpoint:     "127.0.0.1:36380",
		EndpointGRPC: "127.0.0.1:36360",
		EndpointHTTP: "127.0.0.1:36370",
		Queue:        "default",
		Mode:         "",
	}
	cn := n.toClusterNode()

	valueOfN := reflect.ValueOf(n)
	valueOfCN := reflect.ValueOf(*cn)
	for i := 0; i < valueOfN.NumField(); i++ {
		assert.Equal(t, valueOfN.Field(i).String(), valueOfCN.Field(i).String())
	}
}

func TestClusterNodeToNode(t *testing.T) {
	cn := getClusterNode()
	n := cn.toNode()

	valueOfCN := reflect.ValueOf(*cn)
	typeOfCN := reflect.TypeOf(*cn)
	valueOfN := reflect.ValueOf(*n)
	for i := 0; i < valueOfCN.NumField(); i++ {
		fieldType := typeOfCN.Field(i)
		if fieldType.Name == "Scheduler" || fieldType.Name == "Raft" || fieldType.Name == "SchedulerCanStart" {
			continue
		}
		assert.Equal(t, valueOfCN.Field(i).String(), valueOfN.Field(i).String())
	}
}

func TestClusterInit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cn := &ClusterNode{}
	cn.Mode = "HA"

	cn.init(ctx)

	assert.Equal(t, "127.0.0.1:36380", cn.GetEndpointMain())
	assert.Equal(t, "127.0.0.1:36380", cn.Endpoint)
	assert.Equal(t, "127.0.0.1:36360", cn.EndpointGRPC)
	assert.Equal(t, "127.0.0.1:36370", cn.EndpointHTTP)
	assert.Equal(t, "default", cn.Queue)
	assert.NotEmpty(t, cn.NodeMapCopy())
	assert.NotEmpty(t, cn.Raft)
}

func TestClusterRegisterNode(t *testing.T) {
	cn := getClusterNode()

	assert.Len(t, cn.NodeMapCopy(), 0)

	cn.registerNode(cn)

	assert.Len(t, cn.NodeMapCopy(), 1)
}

func TestNodeMapToPbNodesPtr(t *testing.T) {
	cn := getClusterNode()
	cn.registerNode(cn)
	pbNs := cn.NodeMapToPbNodesPtr()

	assert.IsType(t, &pb.Nodes{}, pbNs)
	assert.NotEmpty(t, pbNs)
}

func TestClusterMainNode(t *testing.T) {
	cn := getClusterNode()
	cn.SetEndpointMain("EndpointHA")
	cn.Endpoint = "EndpointTest"
	cn.Mode = "HA"

	cnHA := getClusterNode()
	cnHA.Endpoint = "EndpointHA"

	assert.Empty(t, cn.MainNode())

	cn.registerNode(cn)
	cn.registerNode(cnHA)

	assert.NotEmpty(t, cn.MainNode())
}

func TestClusterHANodeMap(t *testing.T) {
	cn := getClusterNode()

	cnHA := getClusterNode()
	cnHA.Endpoint = "EndpointHA"
	cnHA.Mode = "HA"

	assert.Len(t, cn.HANodeMap(), 0)

	cn.registerNode(cn)
	cn.registerNode(cnHA)

	assert.Len(t, cn.HANodeMap(), 1)
}

func TestClusterChoiceNode(t *testing.T) {
	cn := getClusterNode()
	cn.registerNode(cn)

	_, err := cn.choiceNode([]string{})
	assert.NoError(t, err)
}

func TestClusterChoiceNodeUnhealthy(t *testing.T) {
	cn := getClusterNode()
	cn.registerNode(cn)
	cn.nodeMap[cn.Endpoint]["health"] = false

	_, err := cn.choiceNode([]string{})
	assert.Error(t, err)
}

func TestClusterChoiceNodeQueueNotExist(t *testing.T) {
	cn := getClusterNode()
	cn.registerNode(cn)

	_, err := cn.choiceNode([]string{"other"})
	assert.Error(t, err)
}

func TestClusterHeartbeat(t *testing.T) {
	gob.Register(time.Time{})

	cn := getClusterNode()
	cn.SetEndpointMain("127.0.0.1:36680")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cn.heartbeat(ctx)

	time.Sleep(200 * time.Millisecond)
}

func TestClusterCheckNode(t *testing.T) {
	cn := getClusterNode()
	endpointBak := cn.Endpoint
	endpoint := "test"
	cn.Endpoint = endpoint
	cn.registerNode(cn)
	cn.Endpoint = endpointBak
	cn.nodeMap[endpoint]["last_heartbeat_time"] = time.Now().UTC().Add(-900 * time.Millisecond)
	assert.Equal(t, true, cn.NodeMapCopy()[endpoint]["health"].(bool))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cn.checkNode(ctx)
	time.Sleep(700 * time.Millisecond)

	assert.Equal(t, false, cn.NodeMapCopy()[endpoint]["health"].(bool))

	cn.nodeMap[endpoint]["last_heartbeat_time"] = time.Now().UTC().Add(-6 * time.Minute)
	time.Sleep(700 * time.Millisecond)

	_, ok := cn.NodeMapCopy()[endpoint]
	assert.Equal(t, false, ok)
}

func TestClusterRPCRegister(t *testing.T) {
	cn := getClusterNode()

	assert.Len(t, cn.NodeMapCopy(), 0)

	cn.RPCRegister(cn.toNode(), &Node{})

	assert.Len(t, cn.NodeMapCopy(), 1)
}

func TestClusterRPCHeartbeat(t *testing.T) {
	cn := getClusterNode()

	assert.Len(t, cn.NodeMapCopy(), 0)

	cn.RPCHeartbeat(cn.toNode(), &Node{})

	assert.Len(t, cn.NodeMapCopy(), 1)
}

func TestClusterRegisterNodeRemote(t *testing.T) {
	gob.Register(time.Time{})

	cn := getClusterNode()
	cn.SetEndpointMain("127.0.0.1:36680")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := cn.RegisterNodeRemote(ctx)
	assert.NoError(t, err)

	time.Sleep(300 * time.Millisecond)
}

func TestClusterHeartbeatRemote(t *testing.T) {
	gob.Register(time.Time{})

	cn := getClusterNode()
	cn.SetEndpointMain("127.0.0.1:36680")

	err := cn.heartbeatRemote(context.TODO())
	assert.NoError(t, err)
}
