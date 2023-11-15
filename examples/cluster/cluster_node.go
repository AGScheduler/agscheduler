// 1. go run examples/cluster/cluster_main.go
// 2. go run examples/cluster/cluster_node.go
// 3. go run examples/cluster/cluster_node.go -e 127.0.0.1:36382 -eh 127.0.0.1:36392 -se 127.0.0.1:36362
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

var mainEndpoint = flag.String("me", "127.0.0.1:36380", "Cluster Main Node endpoint")
var endpoint = flag.String("e", "127.0.0.1:36381", "Cluster Node endpoint")
var endpointHTTP = flag.String("eh", "127.0.0.1:36391", "Cluster Node endpoint HTTP")
var schedulerEndpoint = flag.String("se", "127.0.0.1:36361", "Cluster Node Scheduler endpoint")
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

	cservice := &services.ClusterService{Cn: cn}
	err = cservice.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start cluster service: %s", err))
		os.Exit(1)
	}

	select {}
}
