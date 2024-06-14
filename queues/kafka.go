package queues

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	KAFKA_TOPIC = "agscheduler-topic"
	KAFKA_GROUP = "agscheduler-group"
)

// Queue jobs in Kafka.
//
// Producer and consumer must be separated,
// otherwise the offset will be fetched incorrectly.
type KafkaQueue struct {
	Producer *kgo.Client
	Consumer *kgo.Client
	Topic    string

	aCli *kadm.Client

	size int
	jobC chan []byte
}

func (q *KafkaQueue) Init(ctx context.Context) error {
	if q.Topic == "" {
		q.Topic = KAFKA_TOPIC
	}

	q.size = int(math.Abs(float64(q.size)))
	q.jobC = make(chan []byte, q.size)

	q.aCli = kadm.NewClient(q.Producer)

	go q.handleMessage(ctx)

	return nil
}

func (q *KafkaQueue) PushJob(bJ []byte) error {
	aCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	topicD, err := q.aCli.ListTopics(aCtx, q.Topic)
	if err != nil {
		return err
	}

	ps := topicD.TopicsList()[0].Partitions
	psCount := len(ps)
	rand.New(rand.NewSource(time.Now().UnixNano()))
	i := rand.Intn(psCount)
	key := []byte(strconv.Itoa(int(ps[i])))

	record := &kgo.Record{Topic: q.Topic, Key: key, Value: bJ}
	if err := q.Producer.ProduceSync(ctx, record).FirstErr(); err != nil {
		return err
	}

	return nil
}

func (q *KafkaQueue) PullJob() <-chan []byte {
	return q.jobC
}

func (q *KafkaQueue) CountJobs() (int, error) {
	countNewest := 0
	countCommitted := 0
	count := 0

	aCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	newestOffsets, err := q.aCli.ListEndOffsets(aCtx, q.Topic)
	if err != nil {
		return -1, err
	}
	for _, oMap := range newestOffsets.KOffsets() {
		for _, o := range oMap {
			countNewest += int(o.EpochOffset().Offset)
		}
	}

	committedOffsets := q.Consumer.CommittedOffsets()
	for _, oMap := range committedOffsets {
		for _, o := range oMap {
			countCommitted += int(o.Offset)
		}
	}

	count = countNewest - countCommitted

	return count, nil
}

func (q *KafkaQueue) Clear() error {
	defer close(q.jobC)

	aCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_, err := q.aCli.DeleteTopic(aCtx, q.Topic)
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
			fetches := q.Consumer.PollFetches(ctx)
			if errs := fetches.Errors(); len(errs) > 0 {
				slog.Error(fmt.Sprintf("KafkaQueue poll fetches error: `%s`", fmt.Sprint(errs)))
				time.Sleep(1 * time.Second)
				continue
			}

			iter := fetches.RecordIter()
			for !iter.Done() {
				record := iter.Next()
				q.jobC <- record.Value
				q.Consumer.CommitOffsets(ctx, map[string]map[int32]kgo.EpochOffset{
					record.Topic: {
						record.Partition: {
							Offset: record.Offset + 1,
						},
					},
				}, nil)
			}
		}
	}
}
