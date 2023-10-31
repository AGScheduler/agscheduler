package agscheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/gorhill/cronexpr"
	"github.com/stretchr/testify/assert"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"
)

func dryRunScheduler(ctx context.Context, j agscheduler.Job) {}

func runSchedulerPanic(ctx context.Context, j agscheduler.Job) { panic(nil) }

func getSchedulerWithStore() *agscheduler.Scheduler {
	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	return scheduler
}

func getJob() agscheduler.Job {
	agscheduler.RegisterFuncs(dryRunScheduler, runSchedulerPanic)

	return agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "50ms",
		Func:     dryRunScheduler,
	}
}

func getJobWithoutFunc() agscheduler.Job {
	return agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "50ms",
	}
}

func getClusterNode() *agscheduler.ClusterNode {
	return &agscheduler.ClusterNode{
		Id:                "1",
		MainEndpoint:      "127.0.0.1:36364",
		Endpoint:          "127.0.0.1:36364",
		SchedulerEndpoint: "127.0.0.1:36363",
		Queue:             "default",
	}
}

func TestSchedulerSetStore(t *testing.T) {
	store := &stores.MemoryStore{}
	s := &agscheduler.Scheduler{}

	assert.Nil(t, agscheduler.GetStore(s))

	s.SetStore(store)

	assert.NotNil(t, agscheduler.GetStore(s))
}

func TestSchedulerSetClusterNode(t *testing.T) {
	cn := getClusterNode()
	s := &agscheduler.Scheduler{}

	assert.Nil(t, agscheduler.GetClusterNode(s))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.SetClusterNode(ctx, cn)

	assert.NotNil(t, agscheduler.GetClusterNode(s))
}

func TestSchedulerAddJob(t *testing.T) {
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()
	j.Interval = "1s"
	j2 := getJob()

	j, _ = s.AddJob(j)
	s.AddJob(j2)

	assert.Equal(t, agscheduler.STATUS_RUNNING, j.Status)

	time.Sleep(500 * time.Millisecond)
}

func TestSchedulerAddJobDatetime(t *testing.T) {
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()
	j.Type = agscheduler.TYPE_DATETIME
	j.StartAt = "2023-09-22 07:30:08"

	j, _ = s.AddJob(j)

	time.Sleep(50 * time.Millisecond)

	_, err := s.GetJob(j.Id)
	assert.ErrorIs(t, err, agscheduler.JobNotFoundError(j.Id))
}

func TestSchedulerAddJobUnregisteredError(t *testing.T) {
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJobWithoutFunc()

	_, err := s.AddJob(j)
	assert.ErrorIs(t, err, agscheduler.FuncUnregisteredError(""))
}

func TestSchedulerAddJobTimeoutError(t *testing.T) {
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()
	j.Timeout = "errorTimeout"

	_, err := s.AddJob(j)
	assert.Contains(t, err.Error(), "Timeout `"+j.Timeout+"` error")
}

func TestSchedulerRunJobPanic(t *testing.T) {
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()
	j.Func = runSchedulerPanic

	s.AddJob(j)

	time.Sleep(50 * time.Millisecond)
}

func TestSchedulerGetJob(t *testing.T) {
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()

	assert.Empty(t, j.Id)

	j, _ = s.AddJob(j)
	j, _ = s.GetJob(j.Id)

	assert.NotEmpty(t, j.Id)
}

func TestSchedulerGetAllJobs(t *testing.T) {
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()

	js, _ := s.GetAllJobs()
	assert.Len(t, js, 0)

	s.AddJob(j)

	js, _ = s.GetAllJobs()
	assert.Len(t, js, 1)
}

func TestSchedulerUpdateJob(t *testing.T) {
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()

	j, _ = s.AddJob(j)

	interval := "2s"
	j.Interval = interval
	j, _ = s.UpdateJob(j)

	assert.Equal(t, interval, j.Interval)
}

func TestSchedulerDeleteJob(t *testing.T) {
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()

	j, _ = s.AddJob(j)
	s.DeleteJob(j.Id)

	_, err := s.GetJob(j.Id)
	assert.ErrorIs(t, err, agscheduler.JobNotFoundError(j.Id))
}

func TestSchedulerDeleteAllJobs(t *testing.T) {
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()

	s.AddJob(j)
	s.DeleteAllJobs()

	js, _ := s.GetAllJobs()
	assert.Len(t, js, 0)
}

func TestSchedulerPauseJob(t *testing.T) {
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()

	j, _ = s.AddJob(j)

	s.PauseJob(j.Id)
	j, _ = s.GetJob(j.Id)
	assert.Equal(t, agscheduler.STATUS_PAUSED, j.Status)
}

func TestSchedulerPauseJobError(t *testing.T) {
	s := getSchedulerWithStore()
	_, err := s.PauseJob("1")

	assert.Error(t, err)
}

func TestSchedulerResumeJob(t *testing.T) {
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()

	j, _ = s.AddJob(j)

	s.PauseJob(j.Id)
	j, _ = s.GetJob(j.Id)
	assert.Equal(t, agscheduler.STATUS_PAUSED, j.Status)

	s.ResumeJob(j.Id)
	j, _ = s.GetJob(j.Id)
	assert.Equal(t, agscheduler.STATUS_RUNNING, j.Status)
}

func TestSchedulerResumeJobError(t *testing.T) {
	s := getSchedulerWithStore()
	_, err := s.ResumeJob("1")

	assert.Error(t, err)
}

func TestSchedulerRunJob(t *testing.T) {
	s := getSchedulerWithStore()
	j := getJob()

	j, _ = s.AddJob(j)

	s.Stop()

	err := s.RunJob(j)
	assert.NoError(t, err)
}

func TestSchedulerScheduleJobLocal(t *testing.T) {
	cn := getClusterNode()
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.SetClusterNode(ctx, cn)
	s.AddJob(j)

	time.Sleep(500 * time.Millisecond)

	err := s.RunJob(j)
	assert.NoError(t, err)
}

func TestSchedulerScheduleJobRemote(t *testing.T) {
	cn := getClusterNode()
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.SetClusterNode(ctx, cn)
	cn.Id = "1"
	s.AddJob(j)

	time.Sleep(500 * time.Millisecond)

	err := s.RunJob(j)
	assert.NoError(t, err)
}

func TestSchedulerScheduleJobQueueNotExist(t *testing.T) {
	cn := getClusterNode()
	s := getSchedulerWithStore()
	defer s.Stop()
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.SetClusterNode(ctx, cn)
	j.Queues = []string{"other"}
	s.AddJob(j)

	time.Sleep(500 * time.Millisecond)

	err := s.RunJob(j)
	assert.Error(t, err)
}

func TestSchedulerStartAndStop(t *testing.T) {
	s := getSchedulerWithStore()
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}

func TestSchedulerStartOnce(t *testing.T) {
	s := getSchedulerWithStore()
	s.Start()
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}

func TestSchedulerStopOnce(t *testing.T) {
	s := getSchedulerWithStore()
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}

func TestCalcNextRunTimeTimezone(t *testing.T) {
	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "1s",
		Timezone: "America/New_York",
		Status:   agscheduler.STATUS_RUNNING,
	}

	nextRunTimeNew, _ := agscheduler.CalcNextRunTime(j)
	assert.Equal(t, "UTC", nextRunTimeNew.Location().String())
}

func TestCalcNextRunTime(t *testing.T) {
	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "1s",
		Timezone: "America/New_York",
		Status:   agscheduler.STATUS_RUNNING,
	}
	timezone, _ := time.LoadLocation(j.Timezone)

	j.Type = agscheduler.TYPE_DATETIME
	j.StartAt = "2023-09-22 07:30:08"
	nextRunTime, _ := time.ParseInLocation(time.DateTime, j.StartAt, timezone)
	nextRunTimeNew, _ := agscheduler.CalcNextRunTime(j)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0).UTC(), nextRunTimeNew)

	j.Type = agscheduler.TYPE_INTERVAL
	interval := "1s"
	j.Interval = interval
	i, _ := time.ParseDuration(interval)
	nextRunTime = time.Now().In(timezone).Add(i)
	nextRunTimeNew, _ = agscheduler.CalcNextRunTime(j)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0).UTC(), nextRunTimeNew)

	j.Type = agscheduler.TYPE_CRON
	cronExpr := "*/1 * * * *"
	j.CronExpr = cronExpr
	nextRunTime = cronexpr.MustParse(cronExpr).Next(time.Now().In(timezone))
	nextRunTimeNew, _ = agscheduler.CalcNextRunTime(j)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0).UTC(), nextRunTimeNew)

	j.Status = agscheduler.STATUS_PAUSED
	nextRunTimeMax, _ := time.ParseInLocation(time.DateTime, "9999-09-09 09:09:09", timezone)
	nextRunTimeNew, _ = agscheduler.CalcNextRunTime(j)
	assert.Equal(t, time.Unix(nextRunTimeMax.Unix(), 0).UTC(), nextRunTimeNew)

	j.Status = agscheduler.STATUS_RUNNING
	j.Type = "unknown"
	_, err := agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestCalcNextRunTimeTimezoneUnknown(t *testing.T) {
	j := agscheduler.Job{Timezone: "unknown"}

	_, err := agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestCalcNextRunTimeStartAtError(t *testing.T) {
	j := agscheduler.Job{
		Type:    agscheduler.TYPE_DATETIME,
		StartAt: "2023-10-22T07:30:08",
	}

	_, err := agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestCalcNextRunTimeIntervalError(t *testing.T) {
	j := agscheduler.Job{
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "2",
	}

	_, err := agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}
