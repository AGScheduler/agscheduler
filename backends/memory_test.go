package backends

import (
	"testing"

	"github.com/agscheduler/agscheduler"
)

func TestMemoryBackend(t *testing.T) {
	backend := &MemoryBackend{}
	recorder := &agscheduler.Recorder{Backend: backend}

	runTest(t, recorder)
}
