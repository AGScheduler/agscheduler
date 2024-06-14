package queues

import (
	"testing"

	"github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

func TestNsqQueue(t *testing.T) {
	var err error

	tcpAddr := "127.0.0.1:4150"
	httpAddr := "http://127.0.0.1:4151"
	config := nsq.NewConfig()

	testTopic := "agscheduler_test_topic"
	messageHandler := &NsqMessageHandler{}

	producer, err := nsq.NewProducer(tcpAddr, config)
	assert.NoError(t, err)
	err = producer.Ping()
	assert.NoError(t, err)
	defer producer.Stop()

	consumer, err := nsq.NewConsumer(testTopic, testQueue, config)
	assert.NoError(t, err)
	consumer.AddHandler(messageHandler)
	err = consumer.ConnectToNSQD(tcpAddr)
	assert.NoError(t, err)
	defer consumer.Stop()

	nq := &NsqQueue{
		Producer: producer,
		Consumer: consumer,
		Mh:       messageHandler,
		Topic:    testTopic,
		HttpAddr: httpAddr,
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			testQueue: nq,
		},
		WorkersPerQueue: 2,
	}

	runTest(t, brk)
}
