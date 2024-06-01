package queues

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores"
)

var testQueue = "agscheduler_test_queue"

func runQueuesSleep(ctx context.Context, j agscheduler.Job) (result string) {
	time.Sleep(2 * time.Second)
	return
}

func runTest(t *testing.T, brk *agscheduler.Broker) {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: runQueuesSleep},
	)

	sto := &stores.MemoryStore{}
	s := &agscheduler.Scheduler{}
	err := s.SetStore(sto)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(ctx)
	err = s.SetBroker(ctx, brk)
	assert.NoError(t, err)

	for i := range 3 {
		job := agscheduler.Job{
			Name:    "Job" + strconv.Itoa(i+1),
			Type:    agscheduler.JOB_TYPE_DATETIME,
			StartAt: "2023-09-22 07:30:08",
			Func:    runQueuesSleep,
		}
		_, err := s.AddJob(job)
		assert.NoError(t, err)
	}

	ch := brk.Queues[testQueue].PullJob()
	assert.Len(t, ch, 0)

	s.Start()
	time.Sleep(1 * time.Second)
	assert.Len(t, ch, 1)

	time.Sleep(1500 * time.Millisecond)
	assert.Len(t, ch, 0)

	err = s.DeleteAllJobs()
	assert.NoError(t, err)

	s.Stop()

	cancel()
	err = brk.Queues[testQueue].Clear()
	assert.NoError(t, err)
	err = sto.Clear()
	assert.NoError(t, err)
}
