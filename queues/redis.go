package queues

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"runtime/debug"

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

	size       int
	jobC       chan []byte
	cancelFunc context.CancelFunc
}

func (q *RedisQueue) Init() error {
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

	var hmCtx context.Context
	hmCtx, q.cancelFunc = context.WithCancel(ctx)
	go q.handleMessage(hmCtx)

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

func (q *RedisQueue) Clear() error {
	defer close(q.jobC)

	q.cancelFunc()

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
				Count:    int64(10),
				Block:    0,
				NoAck:    false,
			}).Result()
			if err != nil {
				slog.Error(fmt.Sprintf("RedisQueue handle message error: `%s`", err))
				continue
			}
			for _, msg := range messages[0].Messages {
				bJ := []byte(fmt.Sprintf("%v", msg.Values["job"]))
				q.jobC <- bJ
			}
		}
	}
}
