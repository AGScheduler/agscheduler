package stores

import (
	"testing"

	"github.com/redis/go-redis/v9"

	"github.com/kwkwc/agscheduler"
)

func TestRedisStore(t *testing.T) {
	url := "redis://127.0.0.1:6379/0"
	opt, _ := redis.ParseURL(url)
	rdb := redis.NewClient(opt)
	store := &RedisStore{
		RDB:         rdb,
		JobsKey:     "agscheduler.test_jobs",
		RunTimesKey: "agscheduler.test_run_times",
	}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	testAGScheduler(t, scheduler)

	store.Clear()
}
