package stores

import (
	"testing"

	"github.com/kwkwc/agscheduler"
)

func TestMemoryStore(t *testing.T) {
	store := &MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	testAGScheduler(t, scheduler)

	store.Clean()
}
