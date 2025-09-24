// go run examples/queues/memory/main.go

package main

import (
	"github.com/agscheduler/agscheduler"
	eq "github.com/agscheduler/agscheduler/examples/queues"
	"github.com/agscheduler/agscheduler/queues"
)

func main() {
	mq := &queues.MemoryQueue{}
	broker := &agscheduler.Broker{
		Queues: map[string]agscheduler.QueuePkg{
			eq.ExampleQueue: {
				Queue:   mq,
				Workers: 2,
			},
		},
	}

	eq.RunExample(broker)
}
