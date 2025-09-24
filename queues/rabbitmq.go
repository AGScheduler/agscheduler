package queues

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/url"
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
	HttpAddr string
	Username string
	Password string

	ch *amqp.Channel

	size int
	jobC chan []byte
}

func (q *RabbitMQQueue) Name() string {
	return "RabbitMQ"
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

type binding struct {
	Destination string `json:"destination"`
}

func (q *RabbitMQQueue) getExchangeBindings() ([]binding, error) {
	url, err := url.JoinPath(
		q.HttpAddr, fmt.Sprintf("/api/exchanges/%%2f/%s/bindings/source", q.Exchange),
	)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(q.Username, q.Password)

	c := &http.Client{Timeout: 3 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var bindings []binding
	if err := json.NewDecoder(resp.Body).Decode(&bindings); err != nil {
		return nil, err
	}
	return bindings, nil
}

// There is a delay when using the HTTP API,
// so we use `QueueDeclarePassive` to get the number of messages,
// but it lacks the `unacknowledged` data, and when `Messages == 0`, the count will be inaccurate.
func (q *RabbitMQQueue) CountJobs() (int, error) {
	if q.HttpAddr == "" {
		return -1, nil
	}

	count := 0

	bindings, err := q.getExchangeBindings()
	if err != nil {
		return -1, fmt.Errorf("failed to get bindings: %s", err)
	}
	for _, binding := range bindings {
		queueName := binding.Destination
		queue, err := q.ch.QueueDeclarePassive(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			return -1, err
		}
		count += queue.Messages
		// Because of the lack of `unacknowledged` data, when `Messages > 0`,
		// consumers on the same queue are blocking, so the number of consumers needs to be added here.
		if queue.Messages > 0 {
			count += queue.Consumers
		}
	}

	return count, nil
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
	_ = q.ch.Close()

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
		false,   // auto-ack
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
			err := d.Ack(false)
			if err != nil {
				slog.Error(fmt.Sprintf("RabbitMQQueue ack error: `%s`", err))
				time.Sleep(1 * time.Second)
				continue
			}
		}
	}
}
