package backends

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores"
)

func dryRunRecorder(ctx context.Context, j agscheduler.Job) (result string) { return }

func runTest(t *testing.T, rec *agscheduler.Recorder) {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: dryRunRecorder},
	)

	s := &agscheduler.Scheduler{}
	sto := &stores.MemoryStore{}
	err := s.SetStore(sto)
	assert.NoError(t, err)
	err = s.SetRecorder(rec)
	assert.NoError(t, err)

	job := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "2s",
		Func:     dryRunRecorder,
	}
	job, err = s.AddJob(job)
	assert.NoError(t, err)

	job2 := agscheduler.Job{
		Name:    "Job2",
		Type:    agscheduler.JOB_TYPE_DATETIME,
		StartAt: "2023-09-22 07:30:08",
		Func:    dryRunRecorder,
	}
	_, err = s.AddJob(job2)
	assert.NoError(t, err)

	records, total, err := rec.GetRecords(job.Id, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, records, 0)
	assert.Equal(t, 0, int(total))

	s.Start()

	slog.Info("Sleep 5s......\n\n")
	time.Sleep(5 * time.Second)

	records, total, err = rec.GetRecords(job.Id, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, records, 2)
	assert.Equal(t, 2, int(total))
	assert.Equal(t, agscheduler.RECORD_STATUS_COMPLETED, records[0].Status)

	records, total, err = rec.GetRecords(job.Id, 2, 1)
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, 2, int(total))

	records, total, err = rec.GetAllRecords(1, 10)
	assert.NoError(t, err)
	assert.Len(t, records, 3)
	assert.Equal(t, 3, int(total))

	records, total, err = rec.GetAllRecords(2, 2)
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, 3, int(total))
	assert.Equal(t, "Job2", records[0].JobName)
	_, _, err = rec.GetAllRecords(10, 10)
	assert.NoError(t, err)

	err = rec.DeleteRecords(job.Id)
	assert.NoError(t, err)
	records, total, err = rec.GetAllRecords(1, 10)
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, 1, int(total))

	err = rec.DeleteAllRecords()
	assert.NoError(t, err)
	records, total, err = rec.GetAllRecords(1, 10)
	assert.NoError(t, err)
	assert.Len(t, records, 0)
	assert.Equal(t, 0, int(total))

	err = s.DeleteAllJobs()
	assert.NoError(t, err)

	s.Stop()

	err = sto.Clear()
	assert.NoError(t, err)
	err = rec.Clear()
	assert.NoError(t, err)
}
