package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/kwkwc/agscheduler"
	pb "github.com/kwkwc/agscheduler/services/proto"
	"github.com/kwkwc/agscheduler/stores"
)

func TestClusterProxy(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &stores.MemoryStore{}

	cnMain := &agscheduler.ClusterNode{
		MainEndpoint: "127.0.0.1:36380",
		Endpoint:     "127.0.0.1:36380",
		EndpointGRPC: "127.0.0.1:36360",
		EndpointHTTP: "127.0.0.1:36370",
	}
	schedulerMain := &agscheduler.Scheduler{}
	err := schedulerMain.SetStore(store)
	assert.NoError(t, err)
	err = schedulerMain.SetClusterNode(ctx, cnMain)
	assert.NoError(t, err)
	cserviceMain := &ClusterService{Cn: cnMain}
	err = cserviceMain.Start()
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	cnNode := &agscheduler.ClusterNode{
		MainEndpoint: cnMain.Endpoint,
		Endpoint:     "127.0.0.1:36381",
		EndpointGRPC: "127.0.0.1:36361",
		EndpointHTTP: "127.0.0.1:36371",
		Queue:        "node",
	}
	scheduler := &agscheduler.Scheduler{}
	err = scheduler.SetStore(store)
	assert.NoError(t, err)
	err = scheduler.SetClusterNode(ctx, cnNode)
	assert.NoError(t, err)
	cservice := &ClusterService{Cn: cnNode}
	err = cservice.Start()
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	baseUrl := "http://" + cnNode.EndpointHTTP
	resp, err := http.Get(baseUrl + "/scheduler/jobs")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	conn, err := grpc.Dial(cnNode.EndpointGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()
	client := pb.NewSchedulerClient(conn)
	pbJs, err := client.GetAllJobs(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	js := agscheduler.PbJobsPtrToJobs(pbJs)
	assert.Len(t, js, 0)

	err = cserviceMain.Stop()
	assert.NoError(t, err)
	err = cservice.Stop()
	assert.NoError(t, err)
}
