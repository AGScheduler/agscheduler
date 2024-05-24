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
	nsqAddr := "127.0.0.1:4150"
	queue := "default"
	config := nsq.NewConfig()

	messageHandler := &queues.NsqMessageHandler{}

	producer, err := nsq.NewProducer(nsqAddr, config)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create producer: %s", err))
		os.Exit(1)
	}
	err = producer.Ping()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to mq: %s", err))
		os.Exit(1)
	}

	consumer, err := nsq.NewConsumer(queue, queue, config)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create consumer: %s", err))
		os.Exit(1)
	}
	consumer.AddHandler(messageHandler)
	err = consumer.ConnectToNSQD(nsqAddr)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to mq: %s", err))
		os.Exit(1)
	}

	nq := &queues.NsqQueue{
		Producer: producer,
		Consumer: consumer,
		Mh:       messageHandler,
		Topic:    queue,
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			queue: nq,
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
