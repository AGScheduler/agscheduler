package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
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
		// EndpointHTTP:      "127.0.0.1:63637",
		// SchedulerEndpoint: "127.0.0.1:36363",
		// Queue:             "default",
	}
	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler.SetClusterNode(ctx, cnMain)
	cservice := ClusterService{Scheduler: scheduler, Cn: cnMain}
	cservice.Start()

	assert.Len(t, cnMain.QueueMap(), 1)
	cn := &agscheduler.ClusterNode{
		MainEndpoint: cnMain.Endpoint,
		// Endpoint:          "127.0.0.1:36366",
		SchedulerEndpoint: "127.0.0.1:36365",
		Queue:             "node",
	}
	cn.RegisterNodeRemote(context.TODO())
	assert.Len(t, cnMain.QueueMap(), 2)

	resp, _ := http.Get("http://127.0.0.1:63637/cluster/nodes")
	assert.Equal(t, 200, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	rJ := &result{}
	json.Unmarshal(body, &rJ)
	assert.Len(t, rJ.Data.(map[string]any), 2)

	time.Sleep(200 * time.Millisecond)
}
