// go run examples/backends/base.go examples/backends/memory.go

package main

import (
	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/backends"
)

func main() {
	backend := &backends.MemoryBackend{}
	recorder := &agscheduler.Recorder{Backend: backend}

	runExample(recorder)
}
