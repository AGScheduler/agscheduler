package queues

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

var testQueue = "agscheduler_test_queue"

func runQueuesSleep(ctx context.Context, j agscheduler.Job) {
	time.Sleep(1 * time.Second)
}

func testAGScheduler(t *testing.T, s *agscheduler.Scheduler) {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: runQueuesSleep},
	)

	for i := range 3 {
		job := agscheduler.Job{
			Name:    "Job" + strconv.Itoa(i+1),
			Type:    agscheduler.TYPE_DATETIME,
			StartAt: "2023-09-22 07:30:08",
			Func:    runQueuesSleep,
		}
		_, err := s.AddJob(job)
		assert.NoError(t, err)
	}

	brk := agscheduler.GetBroker(s)
	ch := brk.Queues[testQueue].PullJob()
	assert.Len(t, ch, 0)

	s.Start()
	time.Sleep(50 * time.Millisecond)
	assert.Len(t, ch, 1)

	time.Sleep(1 * time.Second)
	assert.Len(t, ch, 0)

	err := s.DeleteAllJobs()
	assert.NoError(t, err)

	s.Stop()
}
