package queues

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"runtime/debug"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	KAFKA_TOPIC = "agscheduler-topic"
)

// Queue jobs in Kafka.
type KafkaQueue struct {
	Cli   *kgo.Client
	Topic string

	size       int
	jobC       chan []byte
	cancelFunc context.CancelFunc
}

func (q *KafkaQueue) Init() error {
	if q.Topic == "" {
		q.Topic = KAFKA_TOPIC
	}

	q.size = int(math.Abs(float64(q.size)))
	q.jobC = make(chan []byte, q.size)

	var hmCtx context.Context
	hmCtx, q.cancelFunc = context.WithCancel(ctx)
	go q.handleMessage(hmCtx)

	return nil
}

func (q *KafkaQueue) PushJob(bJ []byte) error {
	record := &kgo.Record{Topic: q.Topic, Value: bJ}
	if err := q.Cli.ProduceSync(ctx, record).FirstErr(); err != nil {
		return err
	}

	return nil
}

func (q *KafkaQueue) PullJob() <-chan []byte {
	return q.jobC
}

func (q *KafkaQueue) Clear() error {
	defer close(q.jobC)

	q.cancelFunc()

	aCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	aCli := kadm.NewClient(q.Cli)
	_, err := aCli.DeleteTopic(aCtx, q.Topic)
	if err != nil {
		return err
	}

	return nil
}

func (q *KafkaQueue) handleMessage(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("KafkaQueue handle message error: `%s`", err))
			slog.Debug(string(debug.Stack()))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			fetches := q.Cli.PollFetches(ctx)
			if errs := fetches.Errors(); len(errs) > 0 {
				panic(fmt.Sprint(errs))
			}

			iter := fetches.RecordIter()
			for !iter.Done() {
				record := iter.Next()
				q.jobC <- record.Value
			}
		}
	}
}
