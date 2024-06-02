package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/agscheduler/agscheduler"
	pb "github.com/agscheduler/agscheduler/services/proto"
)

func testSchedulerGRPC(t *testing.T, c pb.SchedulerClient) {
	ctx := context.Background()

	_, err := c.Start(ctx, &emptypb.Empty{})
	assert.NoError(t, err)

	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "1s",
		FuncName: "github.com/agscheduler/agscheduler/services.dryRunGRPC",
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	assert.Empty(t, j.Status)

	pbJ, err := agscheduler.JobToPbJobPtr(j)
	assert.NoError(t, err)
	pbJ, err = c.AddJob(ctx, pbJ)
	assert.NoError(t, err)
	j = agscheduler.PbJobPtrToJob(pbJ)
	assert.Equal(t, agscheduler.JOB_STATUS_RUNNING, j.Status)

	j.Type = agscheduler.JOB_TYPE_CRON
	j.CronExpr = "*/1 * * * *"
	pbJ, err = agscheduler.JobToPbJobPtr(j)
	assert.NoError(t, err)
	pbJ, err = c.UpdateJob(ctx, pbJ)
	assert.NoError(t, err)
	j = agscheduler.PbJobPtrToJob(pbJ)
	assert.Equal(t, agscheduler.JOB_TYPE_CRON, j.Type)

	assert.NoError(t, err)
	nextRunTimeMax, err := agscheduler.GetNextRunTimeMax()
	assert.NoError(t, err)

	pbJ, err = c.PauseJob(ctx, &pb.JobReq{Id: j.Id})
	assert.NoError(t, err)
	j = agscheduler.PbJobPtrToJob(pbJ)
	assert.Equal(t, agscheduler.JOB_STATUS_PAUSED, j.Status)
	assert.Equal(t, nextRunTimeMax.Unix(), j.NextRunTime.Unix())

	pbJ, err = c.ResumeJob(ctx, &pb.JobReq{Id: j.Id})
	assert.NoError(t, err)
	j = agscheduler.PbJobPtrToJob(pbJ)
	assert.NotEqual(t, nextRunTimeMax.Unix(), j.NextRunTime.Unix())

	_, err = c.RunJob(ctx, pbJ)
	assert.NoError(t, err)

	_, err = c.ScheduleJob(ctx, pbJ)
	assert.NoError(t, err)

	_, err = c.DeleteJob(ctx, &pb.JobReq{Id: j.Id})
	assert.NoError(t, err)
	_, err = c.GetJob(ctx, &pb.JobReq{Id: j.Id})
	assert.Contains(t, err.Error(), agscheduler.JobNotFoundError(j.Id).Error())

	_, err = c.DeleteAllJobs(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	jsResp, err := c.GetAllJobs(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	js := agscheduler.PbJobsPtrToJobs(jsResp.Jobs)
	assert.Len(t, js, 0)

	_, err = c.Stop(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
}
