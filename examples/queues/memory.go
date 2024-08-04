// go run examples/queues/base.go examples/queues/memory.go

package main

import (
	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/queues"
)

func main() {
	mq := &queues.MemoryQueue{}
	broker := &agscheduler.Broker{
		Queues: map[string]agscheduler.QueuePkg{
			exampleQueue: {
				Queue:   mq,
				Workers: 2,
			},
		},
	}

	runExample(broker)
}
