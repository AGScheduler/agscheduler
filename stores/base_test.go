package stores

import (
	"log"
	"testing"
	"time"

	"github.com/kwkwc/agscheduler"
	"github.com/stretchr/testify/assert"
)

func printMsg(j agscheduler.Job) {
	log.Printf("Run %s %s\n", j.Name, j.Args)
}

func testAGScheduler(t *testing.T, s *agscheduler.Scheduler) {
	agscheduler.RegisterFuncs(printMsg)

	s.Start()

	job := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Func:     printMsg,
		Args:     []any{"arg1", "arg2", "arg3"},
		Interval: 1 * time.Second,
	}
	assert.Equal(t, "", job.FuncName)
	assert.Equal(t, "", job.Status)

	jobId := s.AddJob(job)
	assert.Equal(t, 32, len(jobId))
	job, _ = s.GetJob(jobId)
	assert.Equal(t, agscheduler.STATUS_RUNNING, job.Status)
	assert.NotEqual(t, "", job.FuncName)

	job.Type = agscheduler.TYPE_CRON
	job.CronExpr = "*/1 * * * *"
	s.UpdateJob(job)
	job, _ = s.GetJob(jobId)
	assert.Equal(t, agscheduler.TYPE_CRON, job.Type)

	timezone, _ := time.LoadLocation(job.Timezone)
	nextRunTimeMax, _ := time.ParseInLocation("2006-01-02 15:04:05", "9999-09-09 09:09:09", timezone)

	s.PauseJob(jobId)
	job, _ = s.GetJob(jobId)
	assert.Equal(t, agscheduler.STATUS_PAUSED, job.Status)
	assert.Equal(t, nextRunTimeMax.Unix(), job.NextRunTime.Unix())

	s.ResumeJob(jobId)
	job, _ = s.GetJob(jobId)
	assert.NotEqual(t, nextRunTimeMax.Unix(), job.NextRunTime.Unix())

	s.DeleteJob(jobId)
	_, err := s.GetJob(jobId)
	assert.NotNil(t, err)

	s.DeleteAllJobs()
	jobs, _ := s.GetAllJobs()
	assert.Equal(t, 0, len(jobs))

	time.Sleep(1 * time.Second)

	s.Stop()
}
