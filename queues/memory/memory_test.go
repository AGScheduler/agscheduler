package queues

import (
	"testing"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/internal/adaptertest"
)

func TestMemoryQueue(t *testing.T) {
	mq := &MemoryQueue{}
	broker := &agscheduler.Broker{
		Queues: map[string]agscheduler.QueuePkg{
			testQueue: {
				Queue:   mq,
				Workers: 2,
			},
		},
	}

	adaptertest.RunQueue(t, broker)
}
