package queues

import (
	"fmt"
	"log/slog"
	"math"
	"net/url"
	"runtime/debug"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	MQTT_TOPIC_PREFIX = "$share/agscheduler/"
	MQTT_TOPIC        = "topic"
)

// Queue jobs in MQTT.
type MqttQueue struct {
	Cli         mqtt.Client
	TopicPrefix string
	Topic       string

	size int
	jobC chan []byte
}

func (q *MqttQueue) Init() error {
	if q.TopicPrefix == "" {
		q.TopicPrefix = MQTT_TOPIC_PREFIX
	}
	if q.Topic == "" {
		q.Topic = MQTT_TOPIC
	}

	q.size = int(math.Abs(float64(q.size)))
	q.jobC = make(chan []byte, q.size)

	t, err := url.JoinPath(MQTT_TOPIC_PREFIX, q.Topic)
	if err != nil {
		return err
	}
	if t := q.Cli.Subscribe(t, 2, q.handleMessage); t.Wait() && t.Error() != nil {
		return t.Error()
	}

	return nil
}

func (q *MqttQueue) PushJob(bJ []byte) error {
	if t := q.Cli.Publish(q.Topic, 2, false, bJ); t.Wait() && t.Error() != nil {
		return t.Error()
	}

	return nil
}

func (q *MqttQueue) PullJob() <-chan []byte {
	return q.jobC
}

func (q *MqttQueue) Clear() error {
	defer close(q.jobC)

	return nil
}

func (q *MqttQueue) handleMessage(c mqtt.Client, msg mqtt.Message) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("MqttQueue handle message error: `%s`", err))
			slog.Debug(string(debug.Stack()))
		}
	}()

	q.jobC <- msg.Payload()
}