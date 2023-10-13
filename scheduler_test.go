package agscheduler_test

import (
	"testing"
	"time"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"

	"github.com/gorhill/cronexpr"
	"github.com/stretchr/testify/assert"
)

func getSchedulerWithStore() *agscheduler.Scheduler {
	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	return scheduler
}

func getJob() agscheduler.Job {
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

	id := s.AddJob(j)
	j, _ = s.GetJob(id)

	assert.Equal(t, agscheduler.STATUS_RUNNING, j.Status)

	time.Sleep(100 * time.Millisecond)
}

func TestSchedulerGetJob(t *testing.T) {
	s := getSchedulerWithStore()
	j := getJob()

	id := s.AddJob(j)
	j2, _ := s.GetJob(id)

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

	id := s.AddJob(j)
	j, _ = s.GetJob(id)

	interval := 2 * time.Second
	j.Interval = interval
	s.UpdateJob(j)
	j, _ = s.GetJob(id)

	assert.Equal(t, interval, j.Interval)
}

func TestSchedulerDeleteJob(t *testing.T) {
	s := getSchedulerWithStore()
	j := getJob()

	id := s.AddJob(j)
	s.DeleteJob(id)

	_, err := s.GetJob(id)
	assert.ErrorIs(t, err, agscheduler.JobNotFoundError(id))
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

	id := s.AddJob(j)

	s.PauseJob(id)
	j, _ = s.GetJob(id)
	assert.Equal(t, agscheduler.STATUS_PAUSED, j.Status)
}

func TestSchedulerPauseJobError(t *testing.T) {
	s := getSchedulerWithStore()
	err := s.PauseJob("1")

	assert.Error(t, err)
}

func TestSchedulerResumeJob(t *testing.T) {
	s := getSchedulerWithStore()
	j := getJob()

	id := s.AddJob(j)

	s.PauseJob(id)
	j, _ = s.GetJob(id)
	assert.Equal(t, agscheduler.STATUS_PAUSED, j.Status)

	s.ResumeJob(id)
	j, _ = s.GetJob(id)
	assert.Equal(t, agscheduler.STATUS_RUNNING, j.Status)
}

func TestSchedulerResumeJobError(t *testing.T) {
	s := getSchedulerWithStore()
	err := s.ResumeJob("1")

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
	startAt, _ := time.ParseInLocation("2006-01-02 15:04:05", "2023-09-22 07:30:08", timezone)
	job.StartAt = startAt
	nextRunTime := startAt.In(timezone)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0), agscheduler.CalcNextRunTime(job))

	job.Type = agscheduler.TYPE_INTERVAL
	interval := 1 * time.Second
	job.Interval = interval
	nextRunTime = time.Now().In(timezone).Add(interval)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0), agscheduler.CalcNextRunTime(job))

	job.Type = agscheduler.TYPE_CRON
	cronExpr := "*/1 * * * *"
	job.CronExpr = cronExpr
	nextRunTime = cronexpr.MustParse(cronExpr).Next(time.Now().In(timezone))
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0), agscheduler.CalcNextRunTime(job))

	job.Status = agscheduler.STATUS_PAUSED
	nextRunTimeMax, _ := time.ParseInLocation("2006-01-02 15:04:05", "9999-09-09 09:09:09", timezone)
	assert.Equal(t, time.Unix(nextRunTimeMax.Unix(), 0), agscheduler.CalcNextRunTime(job))

	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()

	job.Status = agscheduler.STATUS_RUNNING
	job.Type = "unknown"
	agscheduler.CalcNextRunTime(job)
}

func TestCalcNextRunTimeTimezoneUnknown(t *testing.T) {
	job := agscheduler.Job{Timezone: "unknown"}

	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()

	agscheduler.CalcNextRunTime(job)
}
