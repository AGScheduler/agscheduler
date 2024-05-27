package queues

import (
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

func TestMqttQueue(t *testing.T) {
	opts := mqtt.NewClientOptions().AddBroker("tcp://127.0.0.1:1883")
	c := mqtt.NewClient(opts)
	token := c.Connect()
	token.Wait()
	assert.NoError(t, token.Error())
	defer c.Disconnect(250)

	mq := &MqttQueue{
		Cli:   c,
		Topic: "test_topic",

		size: 5,
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			testQueue: mq,
		},
		MaxWorkers: 2,
	}

	runTest(t, brk)
}