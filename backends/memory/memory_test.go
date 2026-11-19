package backends

import (
	"testing"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/internal/adaptertest"
)

func TestMemoryBackend(t *testing.T) {
	backend := &MemoryBackend{}
	recorder := &agscheduler.Recorder{Backend: backend}

	adaptertest.RunBackend(t, recorder)
}
