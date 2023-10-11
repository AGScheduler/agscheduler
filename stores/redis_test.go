package stores

import (
	"testing"

	"github.com/redis/go-redis/v9"

	"github.com/kwkwc/agscheduler"
)

func TestRedisStore(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0,
		Password: "",
	})
	store := &RedisStore{RDB: rdb}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	testAGScheduler(t, scheduler)
}
