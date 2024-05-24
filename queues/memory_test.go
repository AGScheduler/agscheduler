package queues

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores"
)

func TestMemoryQueue(t *testing.T) {
	mq := &MemoryQueue{}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			testQueue: mq,
		},
		MaxWorkers: 2,
	}

	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)
	err = scheduler.SetBroker(brk)
	assert.NoError(t, err)

	testAGScheduler(t, scheduler)

	err = store.Clear()
	assert.NoError(t, err)
	err = brk.Queues[testQueue].Clear()
	assert.NoError(t, err)
}
