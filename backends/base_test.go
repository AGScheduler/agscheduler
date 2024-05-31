package backends

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/examples"
	"github.com/agscheduler/agscheduler/stores"
)

func runTest(t *testing.T, rec *agscheduler.Recorder) {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: examples.PrintMsg},
	)

	s := &agscheduler.Scheduler{}
	sto := &stores.MemoryStore{}
	err := s.SetStore(sto)
	assert.NoError(t, err)
	err = s.SetRecorder(rec)
	assert.NoError(t, err)

	job := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "2s",
		Func:     examples.PrintMsg,
	}
	job, err = s.AddJob(job)
	assert.NoError(t, err)

	job2 := agscheduler.Job{
		Name:    "Job",
		Type:    agscheduler.TYPE_DATETIME,
		StartAt: "2023-09-22 07:30:08",
		Func:    examples.PrintMsg,
	}
	_, err = s.AddJob(job2)
	assert.NoError(t, err)

	records, err := rec.GetRecords(job.Id)
	assert.NoError(t, err)
	assert.Len(t, records, 0)

	s.Start()

	slog.Info("Sleep 5s......\n\n")
	time.Sleep(5 * time.Second)

	records, err = rec.GetRecords(job.Id)
	assert.NoError(t, err)
	assert.Len(t, records, 2)
	assert.Equal(t, agscheduler.RECORD_STATUS_COMPLETED, records[0].Status)

	records, err = rec.GetAllRecords()
	assert.NoError(t, err)
	assert.Len(t, records, 3)

	err = rec.DeleteRecords(job.Id)
	assert.NoError(t, err)
	records, err = rec.GetAllRecords()
	assert.NoError(t, err)
	assert.Len(t, records, 1)

	err = rec.DeleteAllRecords()
	assert.NoError(t, err)
	records, err = rec.GetAllRecords()
	assert.NoError(t, err)
	assert.Len(t, records, 0)

	err = s.DeleteAllJobs()
	assert.NoError(t, err)

	s.Stop()

	err = sto.Clear()
	assert.NoError(t, err)
	err = rec.Clear()
	assert.NoError(t, err)
}
