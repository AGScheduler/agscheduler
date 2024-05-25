// go run examples/queues/base.go examples/queues/redis.go

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/queues"
	"github.com/agscheduler/agscheduler/stores"
)

func main() {
	url := "redis://127.0.0.1:6379/1"
	opt, err := redis.ParseURL(url)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to parse url: %s", err))
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to MQ: %s", err))
		os.Exit(1)
	}

	rq := &queues.RedisQueue{
		RDB:      rdb,
		Stream:   "agscheduler_example_stream",
		Group:    "agscheduler_example_group",
		Consumer: "agscheduler_example_consumer",
	}
	brk := &agscheduler.Broker{
		Queues: map[string]agscheduler.Queue{
			exampleQueue: rq,
		},
		MaxWorkers: 2,
	}

	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	err = scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}
	err = scheduler.SetBroker(brk)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set broker: %s", err))
		os.Exit(1)
	}

	runExample(scheduler)
}
