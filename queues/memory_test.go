package queues

import (
	"testing"

	"github.com/agscheduler/agscheduler"
)

func TestMemoryQueue(t *testing.T) {
	mq := &MemoryQueue{}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			testQueue: mq,
		},
		MaxWorkers: 2,
	}

	runTest(t, brk)
}
