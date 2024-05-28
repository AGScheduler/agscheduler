package queues

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"runtime/debug"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	RABBITMQ_EXCHANGE = "agscheduler_exchange"
	RABBITMQ_QUEUE    = "agscheduler_queue"
)

// Queue jobs in RabbitMQ.
type RabbitMQQueue struct {
	Conn     *amqp.Connection
	Exchange string
	Queue    string

	ch *amqp.Channel

	size int
	jobC chan []byte
}

func (q *RabbitMQQueue) Init(ctx context.Context) error {
	if q.Exchange == "" {
		q.Exchange = RABBITMQ_EXCHANGE
	}
	if q.Queue == "" {
		q.Queue = RABBITMQ_QUEUE
	}

	q.size = int(math.Abs(float64(q.size)))
	q.jobC = make(chan []byte, q.size)

	var err error
	q.ch, err = q.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %s", err)
	}
	err = q.ch.ExchangeDeclare(
		q.Exchange, // name
		"fanout",   // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare an exchange: %s", err)
	}
	_, err = q.ch.QueueDeclare(
		q.Queue, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %s", err)
	}
	err = q.ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set Qos: %s", err)
	}
	err = q.ch.QueueBind(
		q.Queue,    // queue name
		"",         // routing key
		q.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind a queue: %s", err)
	}

	go q.handleMessage(ctx)

	return nil
}

func (q *RabbitMQQueue) PushJob(bJ []byte) error {
	pCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := q.ch.PublishWithContext(pCtx,
		q.Exchange, // exchange
		"",         // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         bJ,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (q *RabbitMQQueue) PullJob() <-chan []byte {
	return q.jobC
}

func (q *RabbitMQQueue) Clear() error {
	defer close(q.jobC)

	_, err := q.ch.QueueDelete(q.Queue, false, false, false)
	if err != nil {
		return err
	}
	err = q.ch.ExchangeDelete(q.Exchange, false, false)
	if err != nil {
		return err
	}
	q.ch.Close()

	return nil
}

func (q *RabbitMQQueue) handleMessage(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("RabbitMQQueue handle message error: `%s`", err))
			slog.Debug(string(debug.Stack()))
		}
	}()

	msgs, err := q.ch.Consume(
		q.Queue, // queue
		"",      // consumer
		true,    // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case d := <-msgs:
			q.jobC <- d.Body
		}
	}
}
