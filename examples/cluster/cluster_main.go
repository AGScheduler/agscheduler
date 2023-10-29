// go run examples/cluster/cluster_main.go

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/examples"
	"github.com/kwkwc/agscheduler/services"
	"github.com/kwkwc/agscheduler/stores"
)

var endpoint = flag.String("e", "127.0.0.1:36364", "Cluster Main endpoint")
var schedulerEndpoint = flag.String("se", "127.0.0.1:36363", "Cluster Main Scheduler endpoint")
var schedulerQueue = flag.String("sq", "default", "Cluster Main Scheduler queue")

func main() {
	agscheduler.RegisterFuncs(examples.PrintMsg)

	flag.Parse()

	store := &stores.MemoryStore{}

	cn := &agscheduler.ClusterNode{
		MainEndpoint:      *endpoint,
		Endpoint:          *endpoint,
		SchedulerEndpoint: *schedulerEndpoint,
		SchedulerQueue:    *schedulerQueue,
	}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}
	err = scheduler.SetClusterNode(context.TODO(), cn)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set cluster node: %s", err))
		os.Exit(1)
	}

	rservice := &services.SchedulerRPCService{Scheduler: scheduler}
	crservice := services.ClusterRPCService{
		Srs: rservice,
		Cn:  cn,
	}
	err = crservice.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start service: %s", err))
		os.Exit(1)
	}

	select {}
}
