// go run examples/queues/mqtt/main.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/agscheduler/agscheduler"
	eq "github.com/agscheduler/agscheduler/examples/queues"
	"github.com/agscheduler/agscheduler/queues"
)

func main() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://127.0.0.1:1883")
	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		slog.Error(fmt.Sprintf("Failed to connect to MQ: %s", token.Error()))
		os.Exit(1)
	}
	defer c.Disconnect(250)

	mq := &queues.MqttQueue{
		Cli:         c,
		TopicPrefix: queues.MQTT_TOPIC_PREFIX,
		Topic:       "example_topic",
	}
	broker := &agscheduler.Broker{
		Queues: map[string]agscheduler.QueuePkg{
			eq.ExampleQueue: {
				Queue:   mq,
				Workers: 2,
			},
		},
	}

	eq.RunExample(broker)
}
