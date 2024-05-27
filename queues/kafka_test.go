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
	testGroup := "agscheduler-test-group"
	testTopic := "agscheduler-test-topic"

	seeds := []string{"127.0.0.1:9092"}
	c, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup(testGroup),
		kgo.ConsumeTopics(testTopic),
		kgo.AllowAutoTopicCreation(),
	)
	assert.NoError(t, err)
	defer c.Close()

	aC := kadm.NewClient(c)
	_, err = aC.CreatePartitions(ctx, 1, testTopic)
	assert.NoError(t, err)

	kq := &KafkaQueue{
		Cli:   c,
		Group: testGroup,
		Topic: testTopic,

		size: 5,
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			testQueue: kq,
		},
		MaxWorkers: 2,
	}

	// PS: On any new group, Kafka internally forces a 3s wait.
	// https://github.com/twmb/franz-go/issues/732
	time.Sleep(5 * time.Second)

	runTest(t, brk)
}
