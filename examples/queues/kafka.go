// go run examples/queues/base.go examples/queues/kafka.go

package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/queues"
)

func main() {
	exampleTopic := "agscheduler-example-topic"
	exampleGroup := "agscheduler-example-group"

	seeds := []string{"127.0.0.1:9092"}
	p, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumeTopics(exampleTopic),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to MQ: %s", err))
		os.Exit(1)
	}
	defer p.Close()
	c, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumeTopics(exampleTopic),
		kgo.ConsumerGroup(exampleGroup),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to MQ: %s", err))
		os.Exit(1)
	}
	defer c.Close()

	// Used to ensure that partitions are allocated to consumer.
	// For examples and testing only.
	aC := kadm.NewClient(p)
	_, err = aC.CreatePartitions(ctx, 1, exampleTopic)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create partition: %s", err))
		os.Exit(1)
	}

	kq := &queues.KafkaQueue{
		Producer: p,
		Consumer: c,
		Topic:    exampleTopic,
	}
	broker := &agscheduler.Broker{
		Queues: map[string]agscheduler.QueuePkg{
			exampleQueue: {
				Queue:   kq,
				Workers: 2,
			},
		},
	}

	// PS: On any new group, Kafka internally forces a 3s wait.
	// https://github.com/twmb/franz-go/issues/732
	slog.Info("Sleep 5s......\n\n")
	time.Sleep(5 * time.Second)

	runExample(broker)
}
