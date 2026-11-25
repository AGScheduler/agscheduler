package services

import (
	"context"
	"encoding/gob"
	"net/rpc"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agscheduler/agscheduler"
	stores "github.com/agscheduler/agscheduler/stores/memory"
)

func testClusterRPC(t *testing.T, c *rpc.Client) {
	gob.Register(map[string]any{})

	filters := make(map[string]any)
	var info map[string]any

	err := c.Call("CRPCService.GetInfo", filters, &info)
	assert.NoError(t, err)
	assert.Len(t, info, 5)

	assert.Equal(t, info["version"], agscheduler.Version)
}

func TestClusterRPCServicesAreIsolated(t *testing.T) {
	gob.Register(map[string]any{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	start := func(endpoint, endpointGRPC, endpointHTTP string) *clusterRPCService {
		t.Helper()

		scheduler := &agscheduler.Scheduler{}
		require.NoError(t, scheduler.SetStore(&stores.MemoryStore{}))
		cn := &agscheduler.ClusterNode{
			EndpointMain: endpoint,
			Endpoint:     endpoint,
			EndpointGRPC: endpointGRPC,
			EndpointHTTP: endpointHTTP,
		}
		require.NoError(t, scheduler.SetClusterNode(ctx, cn))

		service := &clusterRPCService{Cn: cn}
		require.NoError(t, service.Start())
		t.Cleanup(func() {
			assert.NoError(t, service.Stop())
		})
		return service
	}

	service1 := start("127.0.0.1:36480", "127.0.0.1:36460", "127.0.0.1:36470")
	service2 := start("127.0.0.1:36481", "127.0.0.1:36461", "127.0.0.1:36471")

	client1, err := rpc.DialHTTP("tcp", service1.Cn.Endpoint)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, client1.Close())
	}()
	client2, err := rpc.DialHTTP("tcp", service2.Cn.Endpoint)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, client2.Close())
	}()

	endpointFrom := func(client *rpc.Client) string {
		t.Helper()

		var info map[string]any
		require.NoError(t, client.Call("CRPCService.GetInfo", map[string]any{}, &info))
		clusterInfo := info["cluster"].(map[string]any)
		mainNode := clusterInfo["main_node"].(map[string]any)
		return mainNode["endpoint"].(string)
	}

	assert.Equal(t, service1.Cn.Endpoint, endpointFrom(client1))
	assert.Equal(t, service2.Cn.Endpoint, endpointFrom(client2))
}
