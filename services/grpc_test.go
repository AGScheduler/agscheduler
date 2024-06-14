package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/backends"
	pb "github.com/agscheduler/agscheduler/services/proto"
	"github.com/agscheduler/agscheduler/stores"
)

func dryRunGRPC(ctx context.Context, j agscheduler.Job) (result string) { return }

func testGRPC(t *testing.T, c pb.BaseClient) {
	ctx := context.Background()

	iResp, err := c.GetInfo(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Len(t, iResp.Info.AsMap(), 5)
	assert.Equal(t, iResp.Info.AsMap()["version"], agscheduler.Version)

	fsResp, err := c.GetFuncs(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	funcLen := len(agscheduler.FuncMap)
	assert.Len(t, fsResp.Funcs, funcLen)
}

func TestGRPCService(t *testing.T) {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: dryRunGRPC},
	)

	scheduler := &agscheduler.Scheduler{}

	store := &stores.MemoryStore{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	mb := &backends.MemoryBackend{}
	recorder := &agscheduler.Recorder{Backend: mb}
	err = scheduler.SetRecorder(recorder)
	assert.NoError(t, err)

	grservice := GRPCService{
		Scheduler: scheduler,
	}
	err = grservice.Start()
	assert.NoError(t, err)

	conn, err := grpc.NewClient(grservice.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	clientB := pb.NewBaseClient(conn)
	testGRPC(t, clientB)
	clientS := pb.NewSchedulerClient(conn)
	testSchedulerGRPC(t, clientS)
	clientR := pb.NewRecorderClient(conn)
	testRecorderGRPC(t, clientS, clientR)

	err = grservice.Stop()
	assert.NoError(t, err)

	err = store.Clear()
	assert.NoError(t, err)
}
