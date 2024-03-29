// go run examples/stores/base.go examples/stores/etcd.go

package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores"
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}
	defer cli.Close()
	store := &stores.EtcdStore{
		Cli:          cli,
		JobsPath:     "/agscheduler/example_jobs",
		RunTimesPath: "/agscheduler/example_run_times",
	}

	scheduler := &agscheduler.Scheduler{}
	err = scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	runExample(scheduler)
}
