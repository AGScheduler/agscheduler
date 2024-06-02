package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/agscheduler/agscheduler"
	pb "github.com/agscheduler/agscheduler/services/proto"
)

func testRecorderGRPC(t *testing.T, sc pb.SchedulerClient, rc pb.RecorderClient) {
	ctx := context.Background()

	_, err := rc.DeleteAllRecords(ctx, &emptypb.Empty{})
	assert.NoError(t, err)

	_, err = sc.Start(ctx, &emptypb.Empty{})
	assert.NoError(t, err)

	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_DATETIME,
		StartAt:  "2023-09-22 07:30:08",
		FuncName: "github.com/agscheduler/agscheduler/services.dryRunGRPC",
	}
	pbJ, err := agscheduler.JobToPbJobPtr(j)
	assert.NoError(t, err)
	pbJ, err = sc.AddJob(ctx, pbJ)
	assert.NoError(t, err)
	j = agscheduler.PbJobPtrToJob(pbJ)

	time.Sleep(1500 * time.Millisecond)

	resp, err := rc.GetRecords(ctx, &pb.RecordsReq{
		JobId: j.Id, Page: int32(1), PageSize: int32(10),
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, int(resp.Total))

	_, err = sc.AddJob(ctx, pbJ)
	assert.NoError(t, err)

	time.Sleep(1500 * time.Millisecond)

	resp, err = rc.GetAllRecords(ctx, &pb.RecordsAllReq{
		Page: int32(1), PageSize: int32(10),
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, int(resp.Total))

	_, err = rc.DeleteRecords(ctx, &pb.JobId{Id: j.Id})
	assert.NoError(t, err)

	resp, err = rc.GetAllRecords(ctx, &pb.RecordsAllReq{
		Page: int32(1), PageSize: int32(10),
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, int(resp.Total))

	_, err = rc.DeleteAllRecords(ctx, &emptypb.Empty{})
	assert.NoError(t, err)

	resp, err = rc.GetAllRecords(ctx, &pb.RecordsAllReq{
		Page: int32(1), PageSize: int32(10),
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, int(resp.Total))

	_, err = sc.Stop(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
}
