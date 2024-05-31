// go run examples/backends/base.go examples/backends/memory.go

package main

import (
	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/backends"
)

func main() {
	mb := &backends.MemoryBackend{}
	recorder := &agscheduler.Recorder{Backend: mb}

	runExample(recorder)
}
