package services

import (
	"context"
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

var ctx = context.Background()

func dryRunRPC(j agscheduler.Job) {}

func testAGSchedulerRPC(t *testing.T, c pb.SchedulerClient) {
	c.Start(ctx, &emptypb.Empty{})

	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "1s",
		FuncName: "github.com/kwkwc/agscheduler/services.dryRunRPC",
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	assert.Empty(t, j.Status)

	pbJ, _ := c.AddJob(ctx, agscheduler.JobToPbJobPtr(j))
	j = agscheduler.PbJobPtrToJob(pbJ)
	assert.Equal(t, agscheduler.STATUS_RUNNING, j.Status)

	j.Type = agscheduler.TYPE_CRON
	j.CronExpr = "*/1 * * * *"
	pbJ, _ = c.UpdateJob(ctx, agscheduler.JobToPbJobPtr(j))
	j = agscheduler.PbJobPtrToJob(pbJ)
	assert.Equal(t, agscheduler.TYPE_CRON, j.Type)

	timezone, _ := time.LoadLocation(j.Timezone)
	nextRunTimeMax, _ := time.ParseInLocation(time.DateTime, "9999-09-09 09:09:09", timezone)

	pbJ, _ = c.PauseJob(ctx, &pb.JobId{Id: j.Id})
	j = agscheduler.PbJobPtrToJob(pbJ)
	assert.Equal(t, agscheduler.STATUS_PAUSED, j.Status)
	assert.Equal(t, nextRunTimeMax.Unix(), j.NextRunTime.Unix())

	pbJ, _ = c.ResumeJob(ctx, &pb.JobId{Id: j.Id})
	j = agscheduler.PbJobPtrToJob(pbJ)
	assert.NotEqual(t, nextRunTimeMax.Unix(), j.NextRunTime.Unix())

	_, err := c.RunJob(ctx, pbJ)
	assert.NoError(t, err)

	c.DeleteJob(ctx, &pb.JobId{Id: j.Id})
	_, err = c.GetJob(ctx, &pb.JobId{Id: j.Id})
	assert.Contains(t, err.Error(), agscheduler.JobNotFoundError(j.Id).Error())

	c.DeleteAllJobs(ctx, &emptypb.Empty{})
	pbJs, _ := c.GetAllJobs(ctx, &emptypb.Empty{})
	js := agscheduler.PbJobsPtrToJobs(pbJs)
	assert.Len(t, js, 0)

	c.Stop(ctx, &emptypb.Empty{})
}

func TestRPCService(t *testing.T) {
	agscheduler.RegisterFuncs(dryRunRPC)

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	srservice := SchedulerRPCService{
		Scheduler: scheduler,
		// Address:   "127.0.0.1:36363",
	}
	srservice.Start()

	conn, _ := grpc.Dial(srservice.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	client := pb.NewSchedulerClient(conn)

	testAGSchedulerRPC(t, client)

	store.Clear()
}
