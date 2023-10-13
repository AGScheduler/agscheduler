// go run base.go redis.go

package main

import (
	"github.com/redis/go-redis/v9"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0,
		Password: "",
	})
	store := &stores.RedisStore{RDB: rdb}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	runExample(scheduler)

	select {}
}
