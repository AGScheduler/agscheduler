package agscheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/gorhill/cronexpr"
	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/backends"
	"github.com/agscheduler/agscheduler/queues"
	"github.com/agscheduler/agscheduler/stores"
)

func dryRunScheduler(ctx context.Context, j agscheduler.Job) (result string) { return }

func runSchedulerPanic(ctx context.Context, j agscheduler.Job) (result string) { panic(nil); return }

func getSchedulerWithStore(t *testing.T) *agscheduler.Scheduler {
	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	return scheduler
}

func getJob() agscheduler.Job {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: dryRunScheduler},
		agscheduler.FuncPkg{Func: runSchedulerPanic},
	)

	return agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "50ms",
		Func:     dryRunScheduler,
	}
}

func getJobWithoutFunc() agscheduler.Job {
	return agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "50ms",
	}
}

func getClusterNode() *agscheduler.ClusterNode {
	return &agscheduler.ClusterNode{
		EndpointMain: "127.0.0.1:36380",
		Endpoint:     "127.0.0.1:36380",
		EndpointGRPC: "127.0.0.1:36360",
		Queue:        "default",
	}
}

func getBroker() *agscheduler.Broker {
	mq := &queues.MemoryQueue{}

	return &agscheduler.Broker{
		Queues: map[string]agscheduler.QueuePkg{
			"default": {
				Queue:   mq,
				Workers: 2,
			},
		},
	}
}

func getRecorder() *agscheduler.Recorder {
	mb := &backends.MemoryBackend{}
	return &agscheduler.Recorder{Backend: mb}
}

func TestSchedulerSetStore(t *testing.T) {
	store := &stores.MemoryStore{}
	s := &agscheduler.Scheduler{}

	assert.Nil(t, agscheduler.GetStore(s))

	err := s.SetStore(store)
	assert.NoError(t, err)

	assert.NotNil(t, agscheduler.GetStore(s))
}

func TestSchedulerSetClusterNode(t *testing.T) {
	cn := getClusterNode()
	s := &agscheduler.Scheduler{}

	assert.Nil(t, agscheduler.GetClusterNode(s))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetClusterNode(ctx, cn)
	assert.NoError(t, err)

	assert.NotNil(t, agscheduler.GetClusterNode(s))
}

func TestSchedulerSetBroker(t *testing.T) {
	brk := &agscheduler.Broker{}
	s := &agscheduler.Scheduler{}

	assert.Nil(t, agscheduler.GetBroker(s))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetBroker(ctx, brk)
	assert.NoError(t, err)

	assert.NotNil(t, agscheduler.GetBroker(s))
}

func TestSchedulerSetRecorder(t *testing.T) {
	rec := getRecorder()
	s := &agscheduler.Scheduler{}

	assert.Nil(t, agscheduler.GetRecorder(s))

	err := s.SetRecorder(rec)
	assert.NoError(t, err)

	assert.NotNil(t, agscheduler.GetRecorder(s))
}

func TestSchedulerAddJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	defer s.Stop()
	j := getJob()
	j.Interval = "1s"
	j2 := getJob()

	j, err := s.AddJob(j)
	assert.NoError(t, err)
	_, err = s.AddJob(j2)
	assert.NoError(t, err)

	assert.Equal(t, agscheduler.JOB_STATUS_RUNNING, j.Status)

	s.Start()
	time.Sleep(500 * time.Millisecond)
}

func TestSchedulerAddJobDatetime(t *testing.T) {
	s := getSchedulerWithStore(t)
	defer s.Stop()
	j := getJob()
	j.Type = agscheduler.JOB_TYPE_DATETIME
	j.StartAt = "2023-09-22 07:30:08"

	j, err := s.AddJob(j)
	assert.NoError(t, err)

	s.Start()
	time.Sleep(50 * time.Millisecond)

	_, err = s.GetJob(j.Id)
	assert.ErrorIs(t, err, agscheduler.JobNotFoundError(j.Id))
}

func TestSchedulerAddJobUnregisteredError(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJobWithoutFunc()

	_, err := s.AddJob(j)
	assert.ErrorIs(t, err, agscheduler.FuncUnregisteredError(""))
}

func TestSchedulerAddJobTimeoutError(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()
	j.Timeout = "errorTimeout"

	_, err := s.AddJob(j)
	assert.Contains(t, err.Error(), "Timeout `"+j.Timeout+"` error")
}

func TestSchedulerRunJobPanic(t *testing.T) {
	s := getSchedulerWithStore(t)
	defer s.Stop()
	j := getJob()
	j.Func = runSchedulerPanic

	_, err := s.AddJob(j)
	assert.NoError(t, err)

	s.Start()
	time.Sleep(50 * time.Millisecond)
}

func TestSchedulerGetJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()

	assert.Empty(t, j.Id)

	j, err := s.AddJob(j)
	assert.NoError(t, err)
	j, err = s.GetJob(j.Id)
	assert.NoError(t, err)

	assert.NotEmpty(t, j.Id)
}

func TestSchedulerGetAllJobs(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()

	js, err := s.GetAllJobs()
	assert.NoError(t, err)
	assert.Len(t, js, 0)

	_, err = s.AddJob(j)
	assert.NoError(t, err)

	js, err = s.GetAllJobs()
	assert.NoError(t, err)
	assert.Len(t, js, 1)
}

func TestSchedulerUpdateJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()

	j, err := s.AddJob(j)
	assert.NoError(t, err)

	interval := "2s"
	j.Interval = interval
	j, err = s.UpdateJob(j)
	assert.NoError(t, err)

	assert.Equal(t, interval, j.Interval)
}

func TestSchedulerDeleteJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()

	j, err := s.AddJob(j)
	assert.NoError(t, err)
	err = s.DeleteJob(j.Id)
	assert.NoError(t, err)

	_, err = s.GetJob(j.Id)
	assert.ErrorIs(t, err, agscheduler.JobNotFoundError(j.Id))
}

func TestSchedulerDeleteAllJobs(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()

	_, err := s.AddJob(j)
	assert.NoError(t, err)
	err = s.DeleteAllJobs()
	assert.NoError(t, err)

	js, err := s.GetAllJobs()
	assert.NoError(t, err)
	assert.Len(t, js, 0)
}

func TestSchedulerPauseJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	defer s.Stop()
	j := getJob()

	j, err := s.AddJob(j)
	assert.NoError(t, err)

	s.Start()

	_, err = s.PauseJob(j.Id)
	assert.NoError(t, err)
	j, err = s.GetJob(j.Id)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.JOB_STATUS_PAUSED, j.Status)
}

func TestSchedulerPauseJobError(t *testing.T) {
	s := getSchedulerWithStore(t)
	_, err := s.PauseJob("1")
	assert.Error(t, err)
}

func TestSchedulerResumeJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	defer s.Stop()
	j := getJob()

	j, err := s.AddJob(j)
	assert.NoError(t, err)

	s.Start()

	_, err = s.PauseJob(j.Id)
	assert.NoError(t, err)
	j, err = s.GetJob(j.Id)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.JOB_STATUS_PAUSED, j.Status)

	_, err = s.ResumeJob(j.Id)
	assert.NoError(t, err)
	j, err = s.GetJob(j.Id)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.JOB_STATUS_RUNNING, j.Status)
}

func TestSchedulerResumeJobError(t *testing.T) {
	s := getSchedulerWithStore(t)
	_, err := s.ResumeJob("1")
	assert.Error(t, err)
}

func TestSchedulerRunJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()

	j, err := s.AddJob(j)
	assert.NoError(t, err)

	s.Stop()

	err = s.RunJob(j)
	assert.NoError(t, err)
}

func TestSchedulerScheduleJobLocal(t *testing.T) {
	cn := getClusterNode()
	s := getSchedulerWithStore(t)
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetClusterNode(ctx, cn)
	assert.NoError(t, err)
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	err = s.ScheduleJob(j)
	assert.NoError(t, err)
}

func TestSchedulerScheduleJobRemote(t *testing.T) {
	cn := getClusterNode()
	s := getSchedulerWithStore(t)
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetClusterNode(ctx, cn)
	assert.NoError(t, err)
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	cn.Endpoint = "test"
	err = s.ScheduleJob(j)
	assert.NoError(t, err)
}

func TestSchedulerScheduleJobClusterQueueNotExist(t *testing.T) {
	cn := getClusterNode()
	s := getSchedulerWithStore(t)
	defer s.Stop()
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetClusterNode(ctx, cn)
	assert.NoError(t, err)
	j.Queues = []string{"other"}
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	s.Start()
	time.Sleep(500 * time.Millisecond)

	err = s.ScheduleJob(j)
	assert.Error(t, err)
}

func TestSchedulerScheduleJobBrokerQueue(t *testing.T) {
	brk := getBroker()
	s := getSchedulerWithStore(t)
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetBroker(ctx, brk)
	assert.NoError(t, err)
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	err = s.ScheduleJob(j)
	assert.NoError(t, err)
}

func TestSchedulerScheduleJobBrokerQueueNotExist(t *testing.T) {
	brk := getBroker()
	brk.Queues = nil
	s := getSchedulerWithStore(t)
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetBroker(ctx, brk)
	assert.NoError(t, err)
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	err = s.ScheduleJob(j)
	assert.Error(t, err)
}

func TestSchedulerBrokerGetQueues(t *testing.T) {
	brk := getBroker()
	qs := brk.GetQueues()
	assert.Len(t, qs, 1)
}

func TestSchedulerRecorderRecordMetadata(t *testing.T) {
	rec := getRecorder()
	s := getSchedulerWithStore(t)
	j := getJob()

	err := s.SetRecorder(rec)
	assert.NoError(t, err)
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	s.Start()
	time.Sleep(500 * time.Millisecond)
}

func TestSchedulerRecorderRecordMetadataTimeout(t *testing.T) {
	rec := getRecorder()
	s := getSchedulerWithStore(t)
	j := getJob()
	j.Timeout = "0.05ms"

	err := s.SetRecorder(rec)
	assert.NoError(t, err)
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	s.Start()
	time.Sleep(500 * time.Millisecond)
}

func TestSchedulerStartAndStop(t *testing.T) {
	s := getSchedulerWithStore(t)
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}

func TestSchedulerStartOnce(t *testing.T) {
	s := getSchedulerWithStore(t)
	s.Start()
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}

func TestSchedulerStopOnce(t *testing.T) {
	s := getSchedulerWithStore(t)
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}

func TestCalcNextRunTimeTimezone(t *testing.T) {
	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "1s",
		Timezone: "America/New_York",
		Status:   agscheduler.JOB_STATUS_RUNNING,
	}

	nextRunTimeNew, err := agscheduler.CalcNextRunTime(j)
	assert.NoError(t, err)
	assert.Equal(t, "UTC", nextRunTimeNew.Location().String())
}

func TestCalcNextRunTime(t *testing.T) {
	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "1s",
		Timezone: "America/New_York",
		Status:   agscheduler.JOB_STATUS_RUNNING,
	}
	timezone, err := time.LoadLocation(j.Timezone)
	assert.NoError(t, err)

	j.Type = agscheduler.JOB_TYPE_DATETIME
	j.StartAt = "2023-09-22 07:30:08"
	nextRunTime, err := time.ParseInLocation(time.DateTime, j.StartAt, timezone)
	assert.NoError(t, err)
	nextRunTimeNew, err := agscheduler.CalcNextRunTime(j)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0).UTC(), nextRunTimeNew)

	j.Type = agscheduler.JOB_TYPE_INTERVAL
	interval := "1s"
	j.Interval = interval
	i, err := time.ParseDuration(interval)
	assert.NoError(t, err)
	nextRunTime = time.Now().In(timezone).Add(i)
	nextRunTimeNew, err = agscheduler.CalcNextRunTime(j)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0).UTC(), nextRunTimeNew)

	j.Type = agscheduler.JOB_TYPE_CRON
	cronExpr := "*/1 * * * *"
	j.CronExpr = cronExpr
	expr, err := cronexpr.Parse(cronExpr)
	assert.NoError(t, err)
	nextRunTime = expr.Next(time.Now().In(timezone))
	nextRunTimeNew, err = agscheduler.CalcNextRunTime(j)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0).UTC(), nextRunTimeNew)

	j.Status = agscheduler.JOB_STATUS_PAUSED
	nextRunTimeMax, err := agscheduler.GetNextRunTimeMax()
	assert.NoError(t, err)
	nextRunTimeNew, err = agscheduler.CalcNextRunTime(j)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(nextRunTimeMax.Unix(), 0).UTC(), nextRunTimeNew)

	j.Status = agscheduler.JOB_STATUS_RUNNING
	j.Type = "unknown"
	_, err = agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestCalcNextRunTimeTimezoneUnknown(t *testing.T) {
	j := agscheduler.Job{Timezone: "unknown"}

	_, err := agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestCalcNextRunTimeStartAtError(t *testing.T) {
	j := agscheduler.Job{
		Type:    agscheduler.JOB_TYPE_DATETIME,
		StartAt: "2023-10-22T07:30:08",
	}

	_, err := agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestCalcNextRunTimeIntervalError(t *testing.T) {
	j := agscheduler.Job{
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "2",
	}

	_, err := agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestCalcNextRunTimeCronExprError(t *testing.T) {
	j := agscheduler.Job{
		Type:     agscheduler.JOB_TYPE_CRON,
		Interval: "*/1 * * * * * * * *",
	}

	_, err := agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestInfo(t *testing.T) {
	s := getSchedulerWithStore(t)
	brk := getBroker()
	rec := getRecorder()
	cn := getClusterNode()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetClusterNode(ctx, cn)
	assert.NoError(t, err)
	err = s.SetBroker(ctx, brk)
	assert.NoError(t, err)
	err = s.SetRecorder(rec)
	assert.NoError(t, err)

	info := s.Info()

	assert.Len(t, info, 5)
	assert.Equal(t, info["version"], agscheduler.Version)
}
