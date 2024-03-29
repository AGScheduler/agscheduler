package stores

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

func TestRedisStore(t *testing.T) {
	url := "redis://127.0.0.1:6379/0"
	opt, err := redis.ParseURL(url)
	assert.NoError(t, err)
	rdb := redis.NewClient(opt)
	defer rdb.Close()
	store := &RedisStore{
		RDB:         rdb,
		JobsKey:     "agscheduler.test_jobs",
		RunTimesKey: "agscheduler.test_run_times",
	}

	scheduler := &agscheduler.Scheduler{}
	err = scheduler.SetStore(store)
	assert.NoError(t, err)

	testAGScheduler(t, scheduler)

	err = store.Clear()
	assert.NoError(t, err)
}
