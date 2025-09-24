// go run examples/stores/redis/main.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"

	es "github.com/agscheduler/agscheduler/examples/stores"
	"github.com/agscheduler/agscheduler/stores"
)

func main() {
	url := "redis://127.0.0.1:6379/0"
	opt, err := redis.ParseURL(url)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to parse url: %s", err))
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)
	_, err = rdb.Ping(es.Ctx).Result()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}
	defer func() {
		_ = rdb.Close()
	}()

	store := &stores.RedisStore{
		RDB:         rdb,
		JobsKey:     "agscheduler.example_jobs",
		RunTimesKey: "agscheduler.example_run_times",
	}

	es.RunExample(store)
}
