package queues

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nsqio/go-nsq"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores"
)

var testTopic = "agscheduler_test_topic"

func TestNsqQueue(t *testing.T) {
	var err error

	addr := "127.0.0.1:4150"
	config := nsq.NewConfig()

	messageHandler := &NsqMessageHandler{}

	producer, err := nsq.NewProducer(addr, config)
	assert.NoError(t, err)
	err = producer.Ping()
	assert.NoError(t, err)

	consumer, err := nsq.NewConsumer(testTopic, testQueue, config)
	assert.NoError(t, err)
	consumer.AddHandler(messageHandler)
	err = consumer.ConnectToNSQD(addr)
	assert.NoError(t, err)

	nq := &NsqQueue{
		Producer: producer,
		Consumer: consumer,
		Mh:       messageHandler,
		Topic:    testTopic,
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			testQueue: nq,
		},
		MaxWorkers: 2,
	}

	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	err = scheduler.SetStore(store)
	assert.NoError(t, err)
	err = scheduler.SetBroker(brk)
	assert.NoError(t, err)

	testAGScheduler(t, scheduler)

	err = store.Clear()
	assert.NoError(t, err)
	err = brk.Queues[testQueue].Clear()
	assert.NoError(t, err)
}
