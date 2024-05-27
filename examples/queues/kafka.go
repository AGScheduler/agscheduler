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
	exampleGroup := "agscheduler-example-group"
	exampleTopic := "agscheduler-example-topic"

	seeds := []string{"127.0.0.1:9092"}
	c, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup(exampleGroup),
		kgo.ConsumeTopics(exampleTopic),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to MQ: %s", err))
		os.Exit(1)
	}
	defer c.Close()

	aC := kadm.NewClient(c)
	_, err = aC.CreatePartitions(ctx, 1, exampleTopic)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create partition: %s", err))
		os.Exit(1)
	}

	kq := &queues.KafkaQueue{
		Cli:   c,
		Group: exampleGroup,
		Topic: exampleTopic,
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			exampleQueue: kq,
		},
		MaxWorkers: 2,
	}

	// PS: On any new group, Kafka internally forces a 3s wait.
	// https://github.com/twmb/franz-go/issues/732
	slog.Info("Sleep 5s......\n\n")
	time.Sleep(5 * time.Second)

	runExample(brk)
}