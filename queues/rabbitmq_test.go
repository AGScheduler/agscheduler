package queues

import (
	"fmt"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

func TestRabbitMQQueue(t *testing.T) {
	username := "guest"
	password := "guest"
	tcpAddr := fmt.Sprintf("amqp://%s:%s@127.0.0.1:5672/", username, password)
	httpAddr := "http://127.0.0.1:15672"

	c, err := amqp.Dial(tcpAddr)
	assert.NoError(t, err)
	defer c.Close()

	rmq := &RabbitMQQueue{
		Conn:     c,
		Exchange: "agscheduler_test_exchange",
		Queue:    testQueue,
		HttpAddr: httpAddr,
		Username: username,
		Password: password,
	}
	broker := &agscheduler.Broker{
		Queues: map[string]agscheduler.QueuePkg{
			testQueue: {
				Queue:   rmq,
				Workers: 2,
			},
		},
	}

	runTest(t, broker)
}
