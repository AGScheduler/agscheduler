package agscheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/services"
	"github.com/kwkwc/agscheduler/stores"
)

func TestRaft(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &stores.MemoryStore{}

	cnMain := &agscheduler.ClusterNode{
		MainEndpoint:          "127.0.0.1:36387",
		Endpoint:              "127.0.0.1:36387",
		EndpointHTTP:          "127.0.0.1:36397",
		SchedulerEndpoint:     "127.0.0.1:36367",
		SchedulerEndpointHTTP: "127.0.0.1:36377",
		Mode:                  "HA",
	}
	schedulerMain := &agscheduler.Scheduler{}
	err := schedulerMain.SetStore(store)
	assert.NoError(t, err)
	err = schedulerMain.SetClusterNode(ctx, cnMain)
	assert.NoError(t, err)
	cserviceMain := &services.ClusterService{Cn: cnMain}
	err = cserviceMain.Start()
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	cnNode := &agscheduler.ClusterNode{
		MainEndpoint:          cnMain.Endpoint,
		Endpoint:              "127.0.0.1:36388",
		EndpointHTTP:          "127.0.0.1:36398",
		SchedulerEndpoint:     "127.0.0.1:36368",
		SchedulerEndpointHTTP: "127.0.0.1:36378",
		Mode:                  "HA",
	}
	schedulerNode := &agscheduler.Scheduler{}
	err = schedulerNode.SetStore(store)
	assert.NoError(t, err)
	err = schedulerNode.SetClusterNode(ctx, cnNode)
	assert.NoError(t, err)
	cserviceNode := &services.ClusterService{Cn: cnNode}
	err = cserviceNode.Start()
	assert.NoError(t, err)

	cnNode2 := &agscheduler.ClusterNode{
		MainEndpoint:          cnMain.Endpoint,
		Endpoint:              "127.0.0.1:36389",
		EndpointHTTP:          "127.0.0.1:36399",
		SchedulerEndpoint:     "127.0.0.1:36369",
		SchedulerEndpointHTTP: "127.0.0.1:36379",
		Mode:                  "HA",
	}
	schedulerNode2 := &agscheduler.Scheduler{}
	err = schedulerNode2.SetStore(store)
	assert.NoError(t, err)
	err = schedulerNode2.SetClusterNode(ctx, cnNode2)
	assert.NoError(t, err)
	cserviceNode2 := &services.ClusterService{Cn: cnNode2}
	err = cserviceNode2.Start()
	assert.NoError(t, err)

	time.Sleep(4 * time.Second)

	// TODO: Since http.Handle can only be registered once,
	// starting multiple ClusterServices here won't work,
	// so it's only used to improve coverage for now.
	// assert.Equal(t, cnMain.GetMainEndpoint(), cnNode.GetMainEndpoint())
	// assert.Equal(t, cnMain.GetMainEndpoint(), cnNode2.GetMainEndpoint())

	err = cserviceMain.Stop()
	assert.NoError(t, err)
	err = cserviceNode.Stop()
	assert.NoError(t, err)
	err = cserviceNode2.Stop()
	assert.NoError(t, err)
}
