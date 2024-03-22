package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/kwkwc/agscheduler"
	pb "github.com/kwkwc/agscheduler/services/proto"
	"github.com/kwkwc/agscheduler/stores"
)

func testGRPC(t *testing.T, c pb.BaseClient) {
	ctx := context.Background()

	pbI, err := c.GetInfo(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Len(t, pbI.Info.AsMap(), 4)

	assert.Equal(t, pbI.Info.AsMap()["version"], agscheduler.Version)
}

func TestGRPCService(t *testing.T) {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: dryRunGRPC},
	)

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	grservice := GRPCService{
		Scheduler: scheduler,
	}
	err = grservice.Start()
	assert.NoError(t, err)

	conn, err := grpc.Dial(grservice.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	clientB := pb.NewBaseClient(conn)
	testGRPC(t, clientB)
	clientS := pb.NewSchedulerClient(conn)
	testSchedulerGRPC(t, clientS)

	err = grservice.Stop()
	assert.NoError(t, err)

	err = store.Clear()
	assert.NoError(t, err)
}
