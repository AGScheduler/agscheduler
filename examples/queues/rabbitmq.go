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
	username := "guest"
	password := "guest"
	tcpAddr := fmt.Sprintf("amqp://%s:%s@127.0.0.1:5672/", username, password)
	httpAddr := "http://127.0.0.1:15672"

	c, err := amqp.Dial(tcpAddr)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to MQ: %s", err))
		os.Exit(1)
	}
	defer c.Close()

	rmq := &queues.RabbitMQQueue{
		Conn:     c,
		Exchange: "agscheduler_example_exchange",
		Queue:    exampleQueue,
		HttpAddr: httpAddr,
		Username: username,
		Password: password,
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			exampleQueue: rmq,
		},
		WorkersPerQueue: 2,
	}

	runExample(brk)
}
