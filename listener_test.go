package agscheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores"
)

func dryRunListener(ctx context.Context, j agscheduler.Job) (result string) { return }

func dryCallbackListener(ep agscheduler.EventPkg) {}

func TestListener(t *testing.T) {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: dryRunListener},
	)

	s := &agscheduler.Scheduler{}

	sto := &stores.MemoryStore{}
	err := s.SetStore(sto)
	assert.NoError(t, err)

	lis := &agscheduler.Listener{
		Callbacks: []agscheduler.CallbackPkg{
			{
				Callback: dryCallbackListener,
				Event:    agscheduler.EVENT_JOB_ADDED | agscheduler.EVENT_JOB_DELETED,
			},
		},
	}
	err = s.SetListener(lis)
	assert.NoError(t, err)

	job := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "2s",
		Func:     dryRunListener,
	}
	job, err = s.AddJob(job)
	assert.NoError(t, err)

	job, err = s.PauseJob(job.Id)
	assert.NoError(t, err)

	err = s.DeleteJob(job.Id)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
}
