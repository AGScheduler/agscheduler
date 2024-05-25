// go run examples/queues/base.go examples/queues/memory.go

package main

import (
	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/queues"
)

func main() {
	mq := &queues.MemoryQueue{}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			exampleQueue: mq,
		},
		MaxWorkers: 2,
	}

	runExample(brk)
}
