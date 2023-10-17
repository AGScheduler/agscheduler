package stores

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/kwkwc/agscheduler"
	"github.com/stretchr/testify/assert"
)

func printMsg(j agscheduler.Job) {
	slog.Info(fmt.Sprintf("Run %s %s\n", j.Name, j.Args))
}

func testAGScheduler(t *testing.T, s *agscheduler.Scheduler) {
	agscheduler.RegisterFuncs(printMsg)

	s.Start()

	time.Sleep(200 * time.Millisecond)

	job := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "1s",
		Func:     printMsg,
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	assert.Empty(t, job.FuncName)
	assert.Empty(t, job.Status)

	job, _ = s.AddJob(job)
	assert.Equal(t, agscheduler.STATUS_RUNNING, job.Status)
	assert.NotEmpty(t, job.FuncName)

	job.Type = agscheduler.TYPE_CRON
	job.CronExpr = "*/1 * * * *"
	job, _ = s.UpdateJob(job)
	assert.Equal(t, agscheduler.TYPE_CRON, job.Type)

	timezone, _ := time.LoadLocation(job.Timezone)
	nextRunTimeMax, _ := time.ParseInLocation(time.DateTime, "9999-09-09 09:09:09", timezone)

	job, _ = s.PauseJob(job.Id)
	assert.Equal(t, agscheduler.STATUS_PAUSED, job.Status)
	assert.Equal(t, nextRunTimeMax.Unix(), job.NextRunTime.Unix())

	job, _ = s.ResumeJob(job.Id)
	assert.NotEqual(t, nextRunTimeMax.Unix(), job.NextRunTime.Unix())

	s.DeleteJob(job.Id)
	_, err := s.GetJob(job.Id)
	assert.ErrorIs(t, err, agscheduler.JobNotFoundError(job.Id))

	s.DeleteAllJobs()
	jobs, _ := s.GetAllJobs()
	assert.Len(t, jobs, 0)

	s.Stop()

	time.Sleep(100 * time.Millisecond)
}
