// go run examples/backends/memory/main.go

package main

import (
	"github.com/agscheduler/agscheduler"
	backends "github.com/agscheduler/agscheduler/backends/memory"
	eb "github.com/agscheduler/agscheduler/examples/backends"
)

func main() {
	backend := &backends.MemoryBackend{}
	recorder := &agscheduler.Recorder{Backend: backend}

	eb.RunExample(recorder)
}
