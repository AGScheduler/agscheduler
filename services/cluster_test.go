package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kwkwc/agscheduler"
	pb "github.com/kwkwc/agscheduler/services/proto"
	"github.com/kwkwc/agscheduler/stores"
)

func TestClusterService(t *testing.T) {
	store := &stores.MemoryStore{}
	cnMain := &agscheduler.ClusterNode{
		MainEndpoint: "127.0.0.1:36380",
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

	assert.Len(t, cnMain.NodeMapCopy(), 1)
	cn := &agscheduler.ClusterNode{
		MainEndpoint: cnMain.Endpoint,
		Endpoint:     "127.0.0.1:36381",
		EndpointGRPC: "127.0.0.1:36361",
		EndpointHTTP: "127.0.0.1:36371",
		Queue:        "node",
	}
	err = cn.RegisterNodeRemote(ctx)
	assert.NoError(t, err)
	assert.Len(t, cnMain.NodeMapCopy(), 2)

	baseUrl := "http://" + cnMain.EndpointHTTP
	testClusterHTTP(t, baseUrl)

	conn, err := grpc.Dial(cnMain.EndpointGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()
	clientC := pb.NewClusterClient(conn)
	testClusterGRPC(t, clientC)

	time.Sleep(200 * time.Millisecond)

	err = cservice.Stop()
	assert.NoError(t, err)
}
