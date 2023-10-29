// 1. go run examples/cluster/cluster_main.go
// 2. go run examples/cluster/cluster_node.go
// 3. go run examples/cluster/cluster_node.go -e 127.0.0.1:36371 -se 127.0.0.1:36370
// 4. go run examples/rpc/rpc_client.go

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

var mainEndpoint = flag.String("me", "127.0.0.1:36364", "Cluster Main endpoint")
var endpoint = flag.String("e", "127.0.0.1:36366", "Cluster Node endpoint")
var schedulerEndpoint = flag.String("se", "127.0.0.1:36365", "Cluster Node Scheduler endpoint")
var schedulerQueue = flag.String("sq", "node", "Cluster Node Scheduler queue")

func main() {
	agscheduler.RegisterFuncs(examples.PrintMsg)

	flag.Parse()

	store := &stores.MemoryStore{}

	cn := &agscheduler.ClusterNode{
		MainEndpoint:      *mainEndpoint,
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

	err = cn.RegisterNodeRemote(context.TODO())
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to register node remote: %s", err))
		os.Exit(1)
	}

	select {}
}
