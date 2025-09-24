package services

import (
	"context"
	"net/rpc"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/agscheduler/agscheduler"
	pb "github.com/agscheduler/agscheduler/services/proto"
	"github.com/agscheduler/agscheduler/stores"
)

func TestClusterService(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	scheduler := &agscheduler.Scheduler{}

	store := &stores.MemoryStore{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	cnMain := &agscheduler.ClusterNode{
		EndpointMain: "127.0.0.1:36380",
	}
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
		EndpointMain: cnMain.Endpoint,
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

	conn, err := grpc.NewClient(cnMain.EndpointGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()
	clientC := pb.NewClusterClient(conn)
	testClusterGRPC(t, clientC)

	rClient, err := rpc.DialHTTP("tcp", cnMain.Endpoint)
	assert.NoError(t, err)
	defer rClient.Close()
	testClusterRPC(t, rClient)

	time.Sleep(200 * time.Millisecond)

	err = cservice.Stop()
	assert.NoError(t, err)
}
