package queues

import (
	"fmt"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler/internal/adaptertest"

	"github.com/agscheduler/agscheduler"
)

func TestKafkaQueue(t *testing.T) {
	suffix := time.Now().UnixNano()
	testTopic := fmt.Sprintf("agscheduler-test-topic-%d", suffix)
	testGroup := fmt.Sprintf("agscheduler-test-group-%d", suffix)

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
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
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

	adaptertest.RunQueue(t, broker)
}
