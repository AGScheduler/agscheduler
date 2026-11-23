package queues

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"runtime/debug"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/agscheduler/agscheduler"
)

const (
	REDIS_STREAM   = "agscheduler_stream"
	REDIS_GROUP    = "agscheduler_group"
	REDIS_CONSUMER = "agscheduler_consumer"
)

// Queue jobs in Redis.
type RedisQueue struct {
	RDB      *redis.Client
	Stream   string
	Group    string
	Consumer string

	size int
	jobC chan agscheduler.QueueMessage
}

func (q *RedisQueue) Name() string {
	return "Redis"
}

func (q *RedisQueue) Init(ctx context.Context) error {
	if q.Stream == "" {
		q.Stream = REDIS_STREAM
	}
	if q.Group == "" {
		q.Group = REDIS_GROUP
	}
	if q.Consumer == "" {
		q.Consumer = REDIS_CONSUMER
	}

	q.size = int(math.Abs(float64(q.size)))
	q.jobC = make(chan agscheduler.QueueMessage, q.size)

	groupIsExist := false
	gs, _ := q.RDB.XInfoGroups(ctx, q.Stream).Result()
	for _, g := range gs {
		if g.Name == q.Group {
			groupIsExist = true
			break
		}
	}
	if !groupIsExist {
		err := q.RDB.XGroupCreateMkStream(ctx, q.Stream, q.Group, "0-0").Err()
		if err != nil {
			return fmt.Errorf("failed to create stream `%s` group `%s`: %s", q.Stream, q.Group, err)
		}
	}

	go q.handleMessage(ctx)

	return nil
}

func (q *RedisQueue) PushJob(bJ []byte) error {
	err := q.RDB.XAdd(ctx, &redis.XAddArgs{
		Stream: q.Stream,
		ID:     "*",
		Values: map[string]any{"job": bJ},
	}).Err()
	if err != nil {
		return err
	}

	return nil
}

func (q *RedisQueue) PullJob() <-chan agscheduler.QueueMessage {
	return q.jobC
}

func (q *RedisQueue) CountJobs() (int, error) {
	count := 0

	gsInfo, err := q.RDB.XInfoGroups(ctx, q.Stream).Result()
	if err != nil {
		return -1, err
	}

	for _, g := range gsInfo {
		count += int(g.Lag + g.Pending)
	}

	return count, nil
}

func (q *RedisQueue) Clear() error {
	defer close(q.jobC)

	err := q.RDB.Del(ctx, q.Stream).Err()
	if err != nil {
		return err
	}

	return nil
}

func (q *RedisQueue) handleMessage(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("RedisQueue handle message error: `%s`", err))
			slog.Debug(string(debug.Stack()))
		}
	}()

	if !q.claimPending(ctx) {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			messages, err := q.RDB.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    q.Group,
				Consumer: q.Consumer,
				Streams:  []string{q.Stream, ">"},
				Count:    int64(1),
				Block:    0,
				NoAck:    false,
			}).Result()
			if err != nil {
				slog.Error(fmt.Sprintf("RedisQueue read group error: `%s`", err))
				time.Sleep(1 * time.Second)
				continue
			}
			for _, msg := range messages[0].Messages {
				if !q.deliver(ctx, msg) {
					return
				}
			}
		}
	}
}

func (q *RedisQueue) claimPending(ctx context.Context) bool {
	start := "0-0"
	for {
		messages, next, err := q.RDB.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   q.Stream,
			Group:    q.Group,
			Consumer: q.Consumer,
			MinIdle:  30 * time.Second,
			Start:    start,
			Count:    100,
		}).Result()
		if err != nil {
			slog.Error(fmt.Sprintf("RedisQueue auto-claim error: `%s`", err))
			return ctx.Err() == nil
		}
		for _, msg := range messages {
			if !q.deliver(ctx, msg) {
				return false
			}
		}
		if next == "0-0" {
			return true
		}
		start = next
	}
}

func (q *RedisQueue) deliver(ctx context.Context, msg redis.XMessage) bool {
	messageID := msg.ID
	bJ := []byte(fmt.Sprintf("%v", msg.Values["job"]))
	delivery := agscheduler.QueueMessage{
		Body: bJ,
		Ack: func() error {
			return q.ack(messageID)
		},
		Nack: func() error {
			return q.requeue(messageID, bJ)
		},
	}
	select {
	case q.jobC <- delivery:
		return true
	case <-ctx.Done():
		return false
	}
}

func (q *RedisQueue) ack(messageID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return q.RDB.XAck(ctx, q.Stream, q.Group, messageID).Err()
}

func (q *RedisQueue) requeue(messageID string, bJ []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := q.RDB.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: q.Stream,
			ID:     "*",
			Values: map[string]any{"job": bJ},
		})
		pipe.XAck(ctx, q.Stream, q.Group, messageID)
		return nil
	})
	return err
}
