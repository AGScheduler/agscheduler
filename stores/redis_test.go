package stores

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisStore(t *testing.T) {
	url := "redis://127.0.0.1:6379/0"
	opt, err := redis.ParseURL(url)
	assert.NoError(t, err)
	rdb := redis.NewClient(opt)
	defer rdb.Close()
	_, err = rdb.Ping(ctx).Result()
	assert.NoError(t, err)

	store := &RedisStore{
		RDB:         rdb,
		JobsKey:     "agscheduler.test_jobs",
		RunTimesKey: "agscheduler.test_run_times",
	}

	runTest(t, store)
}
