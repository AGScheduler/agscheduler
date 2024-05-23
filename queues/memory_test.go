package queues

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores"
)

func TestMemoryQueue(t *testing.T) {
	store := &stores.MemoryStore{}

	mq := &MemoryQueue{}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			"default": mq,
		},
		MaxWorkers: 2,
	}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)
	err = scheduler.SetBroker(brk)
	assert.NoError(t, err)

	testAGScheduler(t, scheduler)

	err = store.Clear()
	assert.NoError(t, err)
	err = brk.Queues["default"].Clear()
	assert.NoError(t, err)
}
