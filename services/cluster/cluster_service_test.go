package services

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

func TestClusterServiceRollsBackStartedServices(t *testing.T) {
	grpcListener := listenTCP(t)
	httpListener := listenTCP(t)
	rpcListener := listenTCP(t)
	defer func() {
		assert.NoError(t, rpcListener.Close())
	}()

	grpcAddress := grpcListener.Addr().String()
	httpAddress := httpListener.Addr().String()
	rpcAddress := rpcListener.Addr().String()
	assert.NoError(t, grpcListener.Close())
	assert.NoError(t, httpListener.Close())

	cn := &agscheduler.ClusterNode{
		EndpointMain: rpcAddress,
		Endpoint:     rpcAddress,
		EndpointGRPC: grpcAddress,
		EndpointHTTP: httpAddress,
		Queue:        "default",
		Scheduler:    &agscheduler.Scheduler{},
	}
	service := &ClusterService{Cn: cn}

	err := service.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cluster RPC Service listen failure")

	assertAddressAvailable(t, grpcAddress)
	assertAddressAvailable(t, httpAddress)
	assert.NoError(t, service.Stop())
}

func listenTCP(t *testing.T) net.Listener {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return listener
}

func assertAddressAvailable(t *testing.T, address string) {
	t.Helper()

	listener, err := net.Listen("tcp", address)
	if assert.NoError(t, err) {
		assert.NoError(t, listener.Close())
	}
}
