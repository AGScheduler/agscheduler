package stores

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kwkwc/agscheduler"
)

func dryRunStores(j agscheduler.Job) {}

func testAGScheduler(t *testing.T, s *agscheduler.Scheduler) {
	agscheduler.RegisterFuncs(dryRunStores)

	s.Start()

	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "1s",
		Func:     dryRunStores,
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	assert.Empty(t, j.FuncName)
	assert.Empty(t, j.Status)

	j, _ = s.AddJob(j)
	assert.Equal(t, agscheduler.STATUS_RUNNING, j.Status)
	assert.NotEmpty(t, j.FuncName)

	j.Type = agscheduler.TYPE_CRON
	j.CronExpr = "*/1 * * * *"
	j, _ = s.UpdateJob(j)
	assert.Equal(t, agscheduler.TYPE_CRON, j.Type)

	timezone, _ := time.LoadLocation(j.Timezone)
	nextRunTimeMax, _ := time.ParseInLocation(time.DateTime, "9999-09-09 09:09:09", timezone)

	j, _ = s.PauseJob(j.Id)
	assert.Equal(t, agscheduler.STATUS_PAUSED, j.Status)
	assert.Equal(t, nextRunTimeMax.Unix(), j.NextRunTime.Unix())

	j, _ = s.ResumeJob(j.Id)
	assert.NotEqual(t, nextRunTimeMax.Unix(), j.NextRunTime.Unix())

	err := s.RunJob(j.Id)
	assert.NoError(t, err)

	s.DeleteJob(j.Id)
	_, err = s.GetJob(j.Id)
	assert.ErrorIs(t, err, agscheduler.JobNotFoundError(j.Id))

	s.DeleteAllJobs()
	js, _ := s.GetAllJobs()
	assert.Len(t, js, 0)

	s.Stop()
}
