// go run examples/queues/base.go examples/queues/memory.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/queues"
	"github.com/agscheduler/agscheduler/stores"
)

func main() {
	mq := &queues.MemoryQueue{}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			"default": mq,
		},
		MaxWorkers: 2,
	}

	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}
	err = scheduler.SetBroker(brk)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set broker: %s", err))
		os.Exit(1)
	}

	runExample(scheduler)
}
