package stores

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kwkwc/agscheduler"
)

func TestMemoryStore(t *testing.T) {
	store := &MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	testAGScheduler(t, scheduler)

	err = store.Clear()
	assert.NoError(t, err)
}
