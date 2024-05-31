package backends

import (
	"testing"

	"github.com/agscheduler/agscheduler"
)

func TestMemoryBackend(t *testing.T) {
	mb := &MemoryBackend{}
	recorder := &agscheduler.Recorder{Backend: mb}

	runTest(t, recorder)
}
