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

var mainEndpoint = flag.String("me", "127.0.0.1:36364", "Cluster Main Node endpoint")
var endpoint = flag.String("e", "127.0.0.1:36366", "Cluster Node endpoint")
var endpointHTTP = flag.String("eh", "127.0.0.1:63638", "Cluster Node endpoint HTTP")
var schedulerEndpoint = flag.String("se", "127.0.0.1:36365", "Cluster Node Scheduler endpoint")
var queue = flag.String("q", "node", "Cluster Node queue")

func main() {
	agscheduler.RegisterFuncs(examples.PrintMsg)

	flag.Parse()

	store := &stores.MemoryStore{}

	cn := &agscheduler.ClusterNode{
		MainEndpoint:      *mainEndpoint,
		Endpoint:          *endpoint,
		EndpointHTTP:      *endpointHTTP,
		SchedulerEndpoint: *schedulerEndpoint,
		Queue:             *queue,
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

	cservice := &services.ClusterService{Scheduler: scheduler, Cn: cn}
	err = cservice.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start cluster service: %s", err))
		os.Exit(1)
	}

	err = cn.RegisterNodeRemote(context.TODO())
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to register node remote: %s", err))
		os.Exit(1)
	}

	select {}
}
