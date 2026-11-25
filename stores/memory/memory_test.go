package stores

import (
	"testing"

	"github.com/agscheduler/agscheduler/internal/adaptertest"
)

func TestMemoryStore(t *testing.T) {
	store := &MemoryStore{}

	adaptertest.RunStore(t, store)
}
