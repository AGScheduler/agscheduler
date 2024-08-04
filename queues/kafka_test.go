package queues

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/agscheduler/agscheduler"
)

func TestKafkaQueue(t *testing.T) {
	testTopic := "agscheduler-test-topic"
	testGroup := "agscheduler-test-group"

	seeds := []string{"127.0.0.1:9092"}
	p, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumeTopics(testTopic),
		kgo.AllowAutoTopicCreation(),
	)
	assert.NoError(t, err)
	defer p.Close()
	c, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumeTopics(testTopic),
		kgo.ConsumerGroup(testGroup),
		kgo.DisableAutoCommit(),
	)
	assert.NoError(t, err)
	defer c.Close()

	// Used to ensure that partitions are allocated to consumer.
	// For examples and testing only.
	aC := kadm.NewClient(p)
	_, err = aC.CreatePartitions(ctx, 1, testTopic)
	assert.NoError(t, err)

	kq := &KafkaQueue{
		Producer: p,
		Consumer: c,
		Topic:    testTopic,
	}
	broker := &agscheduler.Broker{
		Queues: map[string]agscheduler.QueuePkg{
			testQueue: {
				Queue:   kq,
				Workers: 2,
			},
		},
	}

	// PS: On any new group, Kafka internally forces a 3s wait.
	// https://github.com/twmb/franz-go/issues/732
	time.Sleep(5 * time.Second)

	runTest(t, broker)
}
