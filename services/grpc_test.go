package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kwkwc/agscheduler"
	pb "github.com/kwkwc/agscheduler/services/proto"
	"github.com/kwkwc/agscheduler/stores"
)

func TestGRPCService(t *testing.T) {
	agscheduler.RegisterFuncs(dryRunGRPC)

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
	clientS := pb.NewSchedulerClient(conn)
	testSchedulerGRPC(t, clientS)

	err = grservice.Stop()
	assert.NoError(t, err)

	err = store.Clear()
	assert.NoError(t, err)
}
