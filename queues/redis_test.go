package queues

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

func TestRedisQueue(t *testing.T) {
	url := "redis://127.0.0.1:6379/1"
	opt, err := redis.ParseURL(url)
	assert.NoError(t, err)
	rdb := redis.NewClient(opt)
	_, err = rdb.Ping(ctx).Result()
	assert.NoError(t, err)
	defer rdb.Close()

	rq := &RedisQueue{
		RDB:      rdb,
		Stream:   "agscheduler_test_stream",
		Group:    "agscheduler_test_group",
		Consumer: "agscheduler_test_consumer",

		size: 5,
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			testQueue: rq,
		},
		MaxWorkers: 2,
	}

	runTest(t, brk)
}
