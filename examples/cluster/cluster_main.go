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

var endpoint = flag.String("e", "127.0.0.1:36380", "Cluster Main Node endpoint")
var endpointHTTP = flag.String("eh", "127.0.0.1:36390", "Cluster Main Node endpoint HTTP")
var schedulerEndpoint = flag.String("se", "127.0.0.1:36360", "Cluster Main Node Scheduler endpoint")
var schedulerEndpointHTTP = flag.String("seh", "127.0.0.1:36370", "Cluster Main Node Scheduler endpoint HTTP")
var queue = flag.String("q", "default", "Cluster Main Node queue")
var mode = flag.String("m", "", "Cluster Node mode")

func main() {
	agscheduler.RegisterFuncs(examples.PrintMsg)

	flag.Parse()

	store := &stores.MemoryStore{}

	cn := &agscheduler.ClusterNode{
		MainEndpoint:          *endpoint,
		Endpoint:              *endpoint,
		EndpointHTTP:          *endpointHTTP,
		SchedulerEndpoint:     *schedulerEndpoint,
		SchedulerEndpointHTTP: *schedulerEndpointHTTP,
		Queue:                 *queue,
		Mode:                  *mode,
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
