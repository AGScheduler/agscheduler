package services

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	stores "github.com/agscheduler/agscheduler/stores/memory"
)

func TestRaft(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &stores.MemoryStore{}

	schedulerMain := &agscheduler.Scheduler{}
	err := schedulerMain.SetStore(store)
	assert.NoError(t, err)
	cnMain := &agscheduler.ClusterNode{
		EndpointMain: "127.0.0.1:36387",
		Endpoint:     "127.0.0.1:36387",
		EndpointGRPC: "127.0.0.1:36367",
		EndpointHTTP: "127.0.0.1:36377",
		Mode:         "HA",
	}
	err = schedulerMain.SetClusterNode(ctx, cnMain)
	assert.NoError(t, err)
	cserviceMain := &ClusterService{Cn: cnMain}
	err = cserviceMain.Start()
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	schedulerNode := &agscheduler.Scheduler{}
	err = schedulerNode.SetStore(store)
	assert.NoError(t, err)
	cnNode := &agscheduler.ClusterNode{
		EndpointMain: cnMain.Endpoint,
		Endpoint:     "127.0.0.1:36388",
		EndpointGRPC: "127.0.0.1:36368",
		EndpointHTTP: "127.0.0.1:36378",
		Mode:         "HA",
	}
	err = schedulerNode.SetClusterNode(ctx, cnNode)
	assert.NoError(t, err)
	cserviceNode := &ClusterService{Cn: cnNode}
	err = cserviceNode.Start()
	assert.NoError(t, err)

	schedulerNode2 := &agscheduler.Scheduler{}
	err = schedulerNode2.SetStore(store)
	assert.NoError(t, err)
	cnNode2 := &agscheduler.ClusterNode{
		EndpointMain: cnMain.Endpoint,
		Endpoint:     "127.0.0.1:36389",
		EndpointGRPC: "127.0.0.1:36369",
		EndpointHTTP: "127.0.0.1:36379",
		Mode:         "HA",
	}
	err = schedulerNode2.SetClusterNode(ctx, cnNode2)
	assert.NoError(t, err)
	cserviceNode2 := &ClusterService{Cn: cnNode2}
	err = cserviceNode2.Start()
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		mainEndpoint := cnMain.GetEndpointMain()
		return cnNode.GetEndpointMain() == mainEndpoint &&
			cnNode2.GetEndpointMain() == mainEndpoint
	}, 8*time.Second, 100*time.Millisecond)
	assert.Never(t, func() bool {
		mainEndpoint := cnMain.GetEndpointMain()
		return cnNode.GetEndpointMain() != mainEndpoint ||
			cnNode2.GetEndpointMain() != mainEndpoint
	}, time.Second, 100*time.Millisecond)

	err = cserviceMain.Stop()
	assert.NoError(t, err)
	err = cserviceNode.Stop()
	assert.NoError(t, err)
	err = cserviceNode2.Stop()
	assert.NoError(t, err)
}
