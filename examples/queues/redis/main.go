// go run examples/queues/redis/main.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/agscheduler/agscheduler"
	eq "github.com/agscheduler/agscheduler/examples/queues"
	"github.com/agscheduler/agscheduler/queues"
)

func main() {
	url := "redis://127.0.0.1:6379/1"
	opt, err := redis.ParseURL(url)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to parse url: %s", err))
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)
	_, err = rdb.Ping(eq.Ctx).Result()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to MQ: %s", err))
		os.Exit(1)
	}
	defer func() {
		_ = rdb.Close()
	}()

	rq := &queues.RedisQueue{
		RDB:      rdb,
		Stream:   "agscheduler_example_stream",
		Group:    "agscheduler_example_group",
		Consumer: "agscheduler_example_consumer",
	}
	broker := &agscheduler.Broker{
		Queues: map[string]agscheduler.QueuePkg{
			eq.ExampleQueue: {
				Queue:   rq,
				Workers: 2,
			},
		},
	}

	eq.RunExample(broker)
}
