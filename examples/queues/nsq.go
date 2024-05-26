// go run examples/queues/base.go examples/queues/nsq.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/nsqio/go-nsq"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/queues"
)

func main() {
	var err error

	addr := "127.0.0.1:4150"
	config := nsq.NewConfig()

	exampleTopic := "agscheduler_example_topic"
	messageHandler := &queues.NsqMessageHandler{}

	producer, err := nsq.NewProducer(addr, config)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create producer: %s", err))
		os.Exit(1)
	}
	err = producer.Ping()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to MQ: %s", err))
		os.Exit(1)
	}
	defer producer.Stop()

	consumer, err := nsq.NewConsumer(exampleTopic, exampleQueue, config)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create consumer: %s", err))
		os.Exit(1)
	}
	consumer.AddHandler(messageHandler)
	err = consumer.ConnectToNSQD(addr)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to MQ: %s", err))
		os.Exit(1)
	}
	defer consumer.Stop()

	nq := &queues.NsqQueue{
		Producer: producer,
		Consumer: consumer,
		Mh:       messageHandler,
		Topic:    exampleTopic,
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			exampleQueue: nq,
		},
		MaxWorkers: 2,
	}

	runExample(brk)
}
