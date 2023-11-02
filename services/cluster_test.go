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
	agscheduler.RegisterFuncs(dryRunRPC)

	store := &stores.MemoryStore{}
	cnMain := &agscheduler.ClusterNode{
		// MainEndpoint:      "127.0.0.1:36364",
		// Endpoint:          "127.0.0.1:36364",
		EndpointHTTP: "127.0.0.1:63637",
		// SchedulerEndpoint: "127.0.0.1:36363",
		// Queue:             "default",
	}
	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = scheduler.SetClusterNode(ctx, cnMain)
	assert.NoError(t, err)

	cservice := ClusterService{Scheduler: scheduler, Cn: cnMain}
	err = cservice.Start()
	assert.NoError(t, err)

	assert.Len(t, cnMain.QueueMap(), 1)
	cn := &agscheduler.ClusterNode{
		MainEndpoint: cnMain.Endpoint,
		// Endpoint:          "127.0.0.1:36366",
		SchedulerEndpoint: "127.0.0.1:36365",
		Queue:             "node",
	}
	err = cn.RegisterNodeRemote(ctx)
	assert.NoError(t, err)
	assert.Len(t, cnMain.QueueMap(), 2)

	resp, err := http.Get("http://" + cnMain.EndpointHTTP + "/cluster/nodes")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ := &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	assert.Len(t, rJ.Data.(map[string]any), 2)

	var queueMap map[string]map[string]map[string]any
	rClient, err := rpc.DialHTTP("tcp", cnMain.Endpoint)
	assert.NoError(t, err)
	filters := make(map[string]any)
	err = rClient.Call("CRPCService.Nodes", filters, &queueMap)
	assert.NoError(t, err)
	assert.Len(t, queueMap, 2)

	time.Sleep(200 * time.Millisecond)
}
