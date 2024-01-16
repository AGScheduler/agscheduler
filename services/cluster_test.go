package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/rpc"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"
)

func TestClusterService(t *testing.T) {
	store := &stores.MemoryStore{}
	cnMain := &agscheduler.ClusterNode{
		MainEndpoint: "127.0.0.1:36380",
		// Endpoint:              "127.0.0.1:36380",
		EndpointHTTP: "127.0.0.1:36390",
		// SchedulerEndpoint:     "127.0.0.1:36360",
		// SchedulerEndpointHTTP: "127.0.0.1:36370",
		// Queue:                 "default",
	}
	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = scheduler.SetClusterNode(ctx, cnMain)
	assert.NoError(t, err)

	cservice := &ClusterService{Cn: cnMain}
	err = cservice.Start()
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	assert.Len(t, cnMain.NodeMap(), 1)
	cn := &agscheduler.ClusterNode{
		MainEndpoint: cnMain.Endpoint,
		// Endpoint:          "127.0.0.1:36381",
		SchedulerEndpoint:     "127.0.0.1:36361",
		SchedulerEndpointHTTP: "127.0.0.1:36371",
		Queue:                 "node",
	}
	err = cn.RegisterNodeRemote(ctx)
	assert.NoError(t, err)
	assert.Len(t, cnMain.NodeMap(), 2)

	resp, err := http.Get("http://" + cnMain.EndpointHTTP + "/cluster/nodes")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ := &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	assert.Len(t, rJ.Data.(map[string]any), 2)

	var nodeMap map[string]map[string]map[string]any
	rClient, err := rpc.DialHTTP("tcp", cnMain.Endpoint)
	assert.NoError(t, err)
	filters := make(map[string]any)
	err = rClient.Call("CRPCService.Nodes", filters, &nodeMap)
	assert.NoError(t, err)
	assert.Len(t, nodeMap, 2)

	time.Sleep(200 * time.Millisecond)

	err = cservice.Stop()
	assert.NoError(t, err)
}
