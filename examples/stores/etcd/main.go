// go run examples/stores/etcd/main.go

package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	es "github.com/agscheduler/agscheduler/examples/stores"
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
	defer func() {
		_ = cli.Close()
	}()

	store := &stores.EtcdStore{
		Cli:          cli,
		JobsPath:     "/agscheduler/example_jobs",
		RunTimesPath: "/agscheduler/example_run_times",
	}

	es.RunExample(store)
}
