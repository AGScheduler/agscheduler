// go run examples/queues/base.go examples/queues/rabbitmq.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/queues"
)

func main() {
	c, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672/")
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to MQ: %s", err))
		os.Exit(1)
	}
	defer c.Close()

	rmq := &queues.RabbitMQQueue{
		Conn:     c,
		Exchange: "agscheduler_example_exchange",
		Queue:    "agscheduler_example_queue",
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			exampleQueue: rmq,
		},
		MaxWorkers: 2,
	}

	runExample(brk)
}
