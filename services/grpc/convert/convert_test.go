package convert_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/services/grpc/convert"
	pb "github.com/agscheduler/agscheduler/services/proto"
)

func testJob() agscheduler.Job {
	return agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "1s",
		FuncName: "test.func",
		Args:     map[string]any{"name": "value"},
	}
}

func testRecord() agscheduler.Record {
	return agscheduler.Record{
		Id:      1,
		JobId:   "1",
		JobName: "Job",
		Status:  agscheduler.RECORD_STATUS_RUNNING,
		StartAt: time.Unix(1, 0).UTC(),
		EndAt:   time.Unix(2, 0).UTC(),
	}
}

func TestJobConversions(t *testing.T) {
	job := testJob()
	pbJob, err := convert.JobToPbJobPtr(job)
	assert.NoError(t, err)
	assert.IsType(t, &pb.Job{}, pbJob)
	assert.Equal(t, job.Name, pbJob.GetName())

	roundTrip := convert.PbJobPtrToJob(pbJob)
	assert.IsType(t, agscheduler.Job{}, roundTrip)
	assert.Equal(t, job.FuncName, roundTrip.FuncName)
	assert.Equal(t, job.Args, roundTrip.Args)

	pbJobs, err := convert.JobsToPbJobsPtr([]agscheduler.Job{job, job})
	assert.NoError(t, err)
	assert.Len(t, pbJobs, 2)
	assert.Len(t, convert.PbJobsPtrToJobs(pbJobs), 2)
}

func TestRecordConversions(t *testing.T) {
	record := testRecord()
	pbRecord, err := convert.RecordToPbRecordPtr(record)
	assert.NoError(t, err)
	assert.IsType(t, &pb.Record{}, pbRecord)

	roundTrip := convert.PbRecordPtrToRecord(pbRecord)
	assert.Equal(t, record, roundTrip)

	pbRecords, err := convert.RecordsToPbRecordsPtr([]agscheduler.Record{record, record})
	assert.NoError(t, err)
	assert.Len(t, pbRecords, 2)
	assert.Len(t, convert.PbRecordsPtrToRecords(pbRecords), 2)
}

func TestQueueConversions(t *testing.T) {
	queues := []map[string]any{
		{"name": "default", "type": "Memory", "count": 1, "workers": 2},
		{"name": "mail", "type": "NSQ", "count": 2, "workers": 1},
	}
	for _, queue := range queues {
		pbQueue, err := convert.QueueToPbQueuePtr(queue)
		assert.NoError(t, err)
		assert.IsType(t, &pb.Queue{}, pbQueue)
	}

	pbQueues, err := convert.QueuesToPbQueuesPtr(queues)
	assert.NoError(t, err)
	assert.Len(t, pbQueues, 2)
}

func TestNodeMapConversion(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cn := &agscheduler.ClusterNode{}
	scheduler := &agscheduler.Scheduler{}
	assert.NoError(t, scheduler.SetClusterNode(ctx, cn))

	pbNodes := convert.NodeMapToPbNodesPtr(cn)
	assert.IsType(t, map[string]*pb.Node{}, pbNodes)
	assert.NotEmpty(t, pbNodes)
}
