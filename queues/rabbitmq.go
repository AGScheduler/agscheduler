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
	RABBITMQ_QUEUE = "agscheduler_queue"
)

// Queue jobs in RabbitMQ.
type RabbitMQQueue struct {
	Conn  *amqp.Connection
	Queue string

	queueC *amqp.Channel

	size       int
	jobC       chan []byte
	cancelFunc context.CancelFunc
}

func (q *RabbitMQQueue) Init() error {
	if q.Queue == "" {
		q.Queue = RABBITMQ_QUEUE
	}

	q.size = int(math.Abs(float64(q.size)))
	q.jobC = make(chan []byte, q.size)

	var err error
	q.queueC, err = q.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %s", err)
	}
	_, err = q.queueC.QueueDeclare(
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
	err = q.queueC.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set Qos: %s", err)
	}

	var hmCtx context.Context
	hmCtx, q.cancelFunc = context.WithCancel(ctx)
	go q.handleMessage(hmCtx)

	return nil
}

func (q *RabbitMQQueue) PushJob(bJ []byte) error {
	pCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := q.queueC.PublishWithContext(pCtx,
		"",      // exchange
		q.Queue, // routing key
		false,   // mandatory
		false,   // immediate
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

	q.cancelFunc()

	_, err := q.queueC.QueueDelete(q.Queue, false, false, false)
	if err != nil {
		return err
	}
	q.queueC.Close()

	return nil
}

func (q *RabbitMQQueue) handleMessage(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("RabbitMQQueue handle message error: `%s`", err))
			slog.Debug(string(debug.Stack()))
		}
	}()

	msgs, err := q.queueC.Consume(
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
		default:
			for d := range msgs {
				q.jobC <- d.Body
			}
		}
	}
}
