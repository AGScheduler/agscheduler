package queues

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

func TestRabbitMQQueue(t *testing.T) {
	c, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672/")
	assert.NoError(t, err)
	defer c.Close()

	rmq := &RabbitMQQueue{
		Conn:     c,
		Exchange: "agscheduler_test_exchange",
		Queue:    testQueue,

		size: 5,
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			testQueue: rmq,
		},
		WorkersPerQueue: 2,
	}

	runTest(t, brk)
}
