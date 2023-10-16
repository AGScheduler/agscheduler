package agscheduler_test

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"

	"github.com/gorhill/cronexpr"
	"github.com/stretchr/testify/assert"
)

func printMsg(j agscheduler.Job) {
	slog.Info(fmt.Sprintf("Run %s %s\n", j.Name, j.Args))
}

func getSchedulerWithStore() *agscheduler.Scheduler {
	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	return scheduler
}

func getJob() agscheduler.Job {
	agscheduler.RegisterFuncs(printMsg)

	return agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: 500 * time.Millisecond,
		Func:     printMsg,
	}
}

func getJobWithoutFunc() agscheduler.Job {
	return agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: 500 * time.Millisecond,
	}
}

func TestSchedulerSetStore(t *testing.T) {
	store := &stores.MemoryStore{}
	s := &agscheduler.Scheduler{}

	assert.Nil(t, s.Store())

	s.SetStore(store)

	assert.NotNil(t, s.Store())
}

func TestSchedulerAddJob(t *testing.T) {
	s := getSchedulerWithStore()
	j := getJob()

	j, _ = s.AddJob(j)

	assert.Equal(t, agscheduler.STATUS_RUNNING, j.Status)

	time.Sleep(100 * time.Millisecond)
}

func TestSchedulerAddJobError(t *testing.T) {
	s := getSchedulerWithStore()
	j := getJobWithoutFunc()

	_, err := s.AddJob(j)
	assert.ErrorIs(t, err, agscheduler.FuncUnregisteredError(""))
}

func TestSchedulerGetJob(t *testing.T) {
	s := getSchedulerWithStore()
	j := getJob()

	j2, _ := s.AddJob(j)

	assert.NotEqual(t, j, j2)
}

func TestSchedulerGetAllJobs(t *testing.T) {
	s := getSchedulerWithStore()
	j := getJob()

	jobs, _ := s.GetAllJobs()
	assert.Len(t, jobs, 0)

	s.AddJob(j)

	jobs, _ = s.GetAllJobs()
	assert.Len(t, jobs, 1)
}

func TestSchedulerUpdateJob(t *testing.T) {
	s := getSchedulerWithStore()
	j := getJob()

	j, _ = s.AddJob(j)

	interval := 2 * time.Second
	j.Interval = interval
	j, _ = s.UpdateJob(j)

	assert.Equal(t, interval, j.Interval)
}

func TestSchedulerDeleteJob(t *testing.T) {
	s := getSchedulerWithStore()
	j := getJob()

	j, _ = s.AddJob(j)
	s.DeleteJob(j.Id)

	_, err := s.GetJob(j.Id)
	assert.ErrorIs(t, err, agscheduler.JobNotFoundError(j.Id))
}

func TestSchedulerDeleteAllJobs(t *testing.T) {
	s := getSchedulerWithStore()
	j := getJob()

	s.AddJob(j)
	s.DeleteAllJobs()

	jobs, _ := s.GetAllJobs()
	assert.Len(t, jobs, 0)
}

func TestSchedulerPauseJob(t *testing.T) {
	s := getSchedulerWithStore()
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

func TestSchedulerStartAndStop(t *testing.T) {
	s := getSchedulerWithStore()
	s.Start()
	time.Sleep(100 * time.Millisecond)
	s.Stop()
	time.Sleep(100 * time.Millisecond)
}

func TestSchedulerStartOnce(t *testing.T) {
	s := getSchedulerWithStore()
	s.Start()
	time.Sleep(100 * time.Millisecond)
	s.Start()
}

func TestSchedulerStopOnce(t *testing.T) {
	s := getSchedulerWithStore()
	s.Start()
	time.Sleep(100 * time.Millisecond)
	s.Stop()
	time.Sleep(100 * time.Millisecond)
	s.Stop()
}

func TestCalcNextRunTime(t *testing.T) {
	job := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: 1 * time.Second,
		Timezone: "UTC",
		Status:   agscheduler.STATUS_RUNNING,
	}
	timezone, _ := time.LoadLocation(job.Timezone)

	job.Type = agscheduler.TYPE_DATETIME
	startAt, _ := time.ParseInLocation(time.DateTime, "2023-09-22 07:30:08", timezone)
	job.StartAt = startAt
	nextRunTime := startAt.In(timezone)
	nextRunTimeNew, _ := agscheduler.CalcNextRunTime(job)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0), nextRunTimeNew)

	job.Type = agscheduler.TYPE_INTERVAL
	interval := 1 * time.Second
	job.Interval = interval
	nextRunTime = time.Now().In(timezone).Add(interval)
	nextRunTimeNew, _ = agscheduler.CalcNextRunTime(job)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0), nextRunTimeNew)

	job.Type = agscheduler.TYPE_CRON
	cronExpr := "*/1 * * * *"
	job.CronExpr = cronExpr
	nextRunTime = cronexpr.MustParse(cronExpr).Next(time.Now().In(timezone))
	nextRunTimeNew, _ = agscheduler.CalcNextRunTime(job)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0), nextRunTimeNew)

	job.Status = agscheduler.STATUS_PAUSED
	nextRunTimeMax, _ := time.ParseInLocation(time.DateTime, "9999-09-09 09:09:09", timezone)
	nextRunTimeNew, _ = agscheduler.CalcNextRunTime(job)
	assert.Equal(t, time.Unix(nextRunTimeMax.Unix(), 0), nextRunTimeNew)

	job.Status = agscheduler.STATUS_RUNNING
	job.Type = "unknown"
	_, err := agscheduler.CalcNextRunTime(job)
	assert.Error(t, err)
}

func TestCalcNextRunTimeTimezoneUnknown(t *testing.T) {
	job := agscheduler.Job{Timezone: "unknown"}

	_, err := agscheduler.CalcNextRunTime(job)
	assert.Error(t, err)
}
