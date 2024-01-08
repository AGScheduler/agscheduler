package agscheduler

import (
	"context"
	"encoding/gob"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getClusterNode() *ClusterNode {
	return &ClusterNode{
		MainEndpoint:      "127.0.0.1:36380",
		Endpoint:          "127.0.0.1:36380",
		SchedulerEndpoint: "127.0.0.1:36360",
		Queue:             "default",
	}
}

func TestClusterToClusterNode(t *testing.T) {
	n := Node{
		MainEndpoint:      "127.0.0.1:36380",
		Endpoint:          "127.0.0.1:36380",
		SchedulerEndpoint: "127.0.0.1:36360",
		Queue:             "default",
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
		if fieldType.Name == "Scheduler" {
			continue
		}
		assert.Equal(t, valueOfCN.Field(i).String(), valueOfN.Field(i).String())
	}
}

func TestClusterInit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cn := &ClusterNode{}
	cn.init(ctx)

	assert.Equal(t, "127.0.0.1:36380", cn.Endpoint)
	assert.Equal(t, "127.0.0.1:36380", cn.MainEndpoint)
	assert.Equal(t, "127.0.0.1:36390", cn.EndpointHTTP)
	assert.Equal(t, "127.0.0.1:36360", cn.SchedulerEndpoint)
	assert.Equal(t, "default", cn.Queue)
	assert.NotEmpty(t, cn.NodeMap())
}

func TestClusterRegisterNode(t *testing.T) {
	cn := getClusterNode()

	assert.Len(t, cn.NodeMap(), 0)

	cn.registerNode(cn)

	assert.Len(t, cn.NodeMap(), 1)
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
	cn.nodeMap[cn.Queue][cn.Endpoint]["health"] = false

	_, err := cn.choiceNode([]string{})
	assert.Error(t, err)
}

func TestClusterChoiceNodeQueueNotExist(t *testing.T) {
	cn := getClusterNode()
	cn.registerNode(cn)

	_, err := cn.choiceNode([]string{"other"})
	assert.Error(t, err)
}

func TestClusterCheckNode(t *testing.T) {
	cn := getClusterNode()
	endpointBak := cn.Endpoint
	endpoint := "test"
	cn.Endpoint = endpoint
	cn.registerNode(cn)
	cn.Endpoint = endpointBak
	cn.nodeMap[cn.Queue][endpoint]["last_heartbeat_time"] = time.Now().UTC().Add(-600 * time.Millisecond)
	assert.Equal(t, true, cn.NodeMap()[cn.Queue][endpoint]["health"].(bool))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cn.checkNode(ctx)
	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, false, cn.NodeMap()[cn.Queue][endpoint]["health"].(bool))

	cn.nodeMap[cn.Queue][endpoint]["last_heartbeat_time"] = time.Now().UTC().Add(-6 * time.Minute)
	time.Sleep(500 * time.Millisecond)

	_, ok := cn.NodeMap()[cn.Queue][endpoint]
	assert.Equal(t, false, ok)
}

func TestClusterRPCRegister(t *testing.T) {
	cn := getClusterNode()

	assert.Len(t, cn.NodeMap(), 0)

	cn.RPCRegister(cn.toNode(), &Node{})

	assert.Len(t, cn.NodeMap(), 1)
}

func TestClusterRPCPing(t *testing.T) {
	cn := getClusterNode()

	assert.Len(t, cn.NodeMap(), 0)

	cn.RPCPing(cn.toNode(), &Node{})

	assert.Len(t, cn.NodeMap(), 1)
}

func TestClusterRegisterNodeRemote(t *testing.T) {
	gob.Register(time.Time{})

	cn := getClusterNode()
	cn.MainEndpoint = "127.0.0.1:36680"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := cn.RegisterNodeRemote(ctx)
	assert.NoError(t, err)

	time.Sleep(300 * time.Millisecond)
}

func TestClusterHeartbeatRemote(t *testing.T) {
	gob.Register(time.Time{})

	cn := getClusterNode()
	cn.MainEndpoint = "127.0.0.1:36680"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cn.heartbeatRemote(ctx)

	time.Sleep(200 * time.Millisecond)
}

func TestClusterPingRemote(t *testing.T) {
	gob.Register(time.Time{})

	cn := getClusterNode()
	cn.MainEndpoint = "127.0.0.1:36680"

	err := cn.pingRemote(context.TODO())
	assert.NoError(t, err)
}
