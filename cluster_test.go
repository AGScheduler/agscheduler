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
		Id:                "1",
		MainEndpoint:      "127.0.0.1:36364",
		Endpoint:          "127.0.0.1:36364",
		SchedulerEndpoint: "127.0.0.1:36363",
		Queue:             "default",
	}
}

func TestClusterToClusterNode(t *testing.T) {
	n := Node{
		Id:                "1",
		MainEndpoint:      "127.0.0.1:36364",
		Endpoint:          "127.0.0.1:36364",
		SchedulerEndpoint: "127.0.0.1:36363",
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
	valueOfN := reflect.ValueOf(*n)
	for i := 0; i < valueOfCN.NumField(); i++ {
		assert.Equal(t, valueOfCN.Field(i).String(), valueOfN.Field(i).String())
	}
}

func TestClusterSetId(t *testing.T) {
	cn := getClusterNode()
	cn.setId()

	assert.Len(t, cn.Id, 16)
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
	cn.nodeMap[cn.Queue][cn.Id]["health"] = false

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
	id := "2"
	cn.Id = id
	cn.registerNode(cn)
	cn.Id = "1"
	cn.nodeMap[cn.Queue][id]["last_heartbeat_time"] = time.Now().UTC().Add(-600 * time.Millisecond)
	assert.Equal(t, true, cn.nodeMap[cn.Queue][id]["health"].(bool))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cn.checkNode(ctx)
	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, false, cn.nodeMap[cn.Queue][id]["health"].(bool))

	cn.nodeMap[cn.Queue][id]["last_heartbeat_time"] = time.Now().UTC().Add(-6 * time.Minute)
	time.Sleep(500 * time.Millisecond)

	_, ok := cn.nodeMap[cn.Queue][id]
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
	cn.MainEndpoint = "127.0.0.1:36664"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := cn.RegisterNodeRemote(ctx)
	assert.NoError(t, err)

	time.Sleep(300 * time.Millisecond)
}

func TestClusterHeartbeatRemote(t *testing.T) {
	gob.Register(time.Time{})

	cn := getClusterNode()
	cn.MainEndpoint = "127.0.0.1:36664"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cn.heartbeatRemote(ctx)

	time.Sleep(200 * time.Millisecond)
}

func TestClusterPingRemote(t *testing.T) {
	gob.Register(time.Time{})

	cn := getClusterNode()
	cn.MainEndpoint = "127.0.0.1:36664"

	err := cn.pingRemote(context.TODO())
	assert.NoError(t, err)
}
