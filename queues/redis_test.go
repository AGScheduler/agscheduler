package queues

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/redis/go-redis/v9"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores"
)

func TestRedisQueue(t *testing.T) {
	url := "redis://127.0.0.1:6379/1"
	opt, err := redis.ParseURL(url)
	assert.NoError(t, err)
	rdb := redis.NewClient(opt)
	defer rdb.Close()
	_, err = rdb.Ping(ctx).Result()
	assert.NoError(t, err)

	rq := &RedisQueue{
		RDB:      rdb,
		Stream:   "agscheduler_test_stream",
		Group:    "agscheduler_test_group",
		Consumer: "agscheduler_test_consumer",
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			testQueue: rq,
		},
		MaxWorkers: 2,
	}

	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	err = scheduler.SetStore(store)
	assert.NoError(t, err)
	err = scheduler.SetBroker(brk)
	assert.NoError(t, err)

	testAGScheduler(t, scheduler)

	err = store.Clear()
	assert.NoError(t, err)
	err = brk.Queues[testQueue].Clear()
	assert.NoError(t, err)
}
