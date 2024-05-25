package queues

import (
	"fmt"
	"log/slog"

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

	jobC chan []byte
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

	q.jobC = make(chan []byte, 5)

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
			return err
		}
	}

	go q.handleMessage()

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

	err := q.RDB.Del(ctx, q.Stream).Err()
	if err != nil {
		return err
	}

	return nil
}

func (q *RedisQueue) handleMessage() {
	for {
		messages, err := q.RDB.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    q.Group,
			Consumer: q.Consumer,
			Streams:  []string{q.Stream, ">"},
			Count:    int64(10),
			Block:    0,
			NoAck:    false,
		}).Result()
		if err != nil {
			slog.Error(fmt.Sprintf("RedisQueue read group error: `%s`", err))
			continue
		}
		for _, msg := range messages[0].Messages {
			bJ := []byte(fmt.Sprintf("%v", msg.Values["job"]))
			q.jobC <- bJ
		}
	}
}
