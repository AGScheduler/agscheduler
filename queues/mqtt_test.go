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
		Cli:         c,
		TopicPrefix: MQTT_TOPIC_PREFIX,
		Topic:       "test_topic",
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.QueuePkg{
			testQueue: {
				Queue:   mq,
				Workers: 2,
			},
		},
	}

	runTest(t, brk)
}
