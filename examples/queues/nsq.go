// go run examples/queues/base.go examples/queues/nsq.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/nsqio/go-nsq"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/queues"
	"github.com/agscheduler/agscheduler/stores"
)

func main() {
	var err error

	addr := "127.0.0.1:4150"
	config := nsq.NewConfig()

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

	consumer, err := nsq.NewConsumer(exampleQueue, exampleQueue, config)
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

	nq := &queues.NsqQueue{
		Producer: producer,
		Consumer: consumer,
		Mh:       messageHandler,
		Topic:    exampleQueue,
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			exampleQueue: nq,
		},
		MaxWorkers: 2,
	}

	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	err = scheduler.SetStore(store)
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
