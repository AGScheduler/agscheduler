// go run examples/queues/nsq/main.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/nsqio/go-nsq"

	"github.com/agscheduler/agscheduler"
	eq "github.com/agscheduler/agscheduler/examples/queues"
	"github.com/agscheduler/agscheduler/queues"
)

func main() {
	var err error

	tcpAddr := "127.0.0.1:4150"
	httpAddr := "http://127.0.0.1:4151"
	config := nsq.NewConfig()

	exampleTopic := "agscheduler_example_topic"
	messageHandler := &queues.NsqMessageHandler{}

	producer, err := nsq.NewProducer(tcpAddr, config)
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

	consumer, err := nsq.NewConsumer(exampleTopic, eq.ExampleQueue, config)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create consumer: %s", err))
		os.Exit(1)
	}
	consumer.AddHandler(messageHandler)
	err = consumer.ConnectToNSQD(tcpAddr)
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
		HttpAddr: httpAddr,
	}
	broker := &agscheduler.Broker{
		Queues: map[string]agscheduler.QueuePkg{
			eq.ExampleQueue: {
				Queue:   nq,
				Workers: 2,
			},
		},
	}

	eq.RunExample(broker)
}
