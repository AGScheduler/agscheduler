package queues

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"runtime/debug"
	"time"

	"github.com/redis/go-redis/v9"
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
	jobC chan []byte
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
	q.jobC = make(chan []byte, q.size)

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

func (q *RedisQueue) PullJob() <-chan []byte {
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
				bJ := []byte(fmt.Sprintf("%v", msg.Values["job"]))
				q.jobC <- bJ
				err := q.RDB.XAck(ctx, q.Stream, q.Group, msg.ID).Err()
				if err != nil {
					slog.Error(fmt.Sprintf("RedisQueue ack error: `%s`", err))
					time.Sleep(1 * time.Second)
					continue
				}
			}
		}
	}
}
