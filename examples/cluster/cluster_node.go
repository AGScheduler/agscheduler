// Normal Mode
//
// 1. go run examples/cluster/cluster_node.go -e 127.0.0.1:36380
// 2. go run examples/cluster/cluster_node.go -em 127.0.0.1:36380 -e 127.0.0.1:36381 -egr 127.0.0.1:36361 -eh 127.0.0.1:36371
// 3. go run examples/cluster/cluster_node.go -em 127.0.0.1:36380 -e 127.0.0.1:36382 -egr 127.0.0.1:36362 -eh 127.0.0.1:36372
// 4. go run examples/grpc/grpc_client.go

// HA Mode
// NOTE: All HA nodes need to connect to the same Store (excluding `MemoryStore`)
//
// 1. go run examples/cluster/cluster_node.go -e 127.0.0.1:36380 -m HA
// 2. go run examples/cluster/cluster_node.go -em 127.0.0.1:36380 -e 127.0.0.1:36381 -egr 127.0.0.1:36361 -eh 127.0.0.1:36371 -m HA
// 3. go run examples/cluster/cluster_node.go -em 127.0.0.1:36380 -e 127.0.0.1:36382 -egr 127.0.0.1:36362 -eh 127.0.0.1:36372 -m HA
// 4. go run examples/grpc/grpc_client.go

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

var endpointMain = flag.String("em", "", "Cluster Main Node endpoint")
var endpoint = flag.String("e", "127.0.0.1:36380", "Cluster Node endpoint")
var endpointGRPC = flag.String("egr", "127.0.0.1:36360", "Cluster Node endpoint gRPC")
var endpointHTTP = flag.String("eh", "127.0.0.1:36370", "Cluster Node endpoint HTTP")
var queue = flag.String("q", "default", "Cluster Node queue")
var mode = flag.String("m", "", "Cluster Node mode, options `HA`")

func main() {
	agscheduler.RegisterFuncs(examples.PrintMsg)

	flag.Parse()

	store := &stores.MemoryStore{}

	cn := &agscheduler.ClusterNode{
		EndpointMain: *endpointMain,
		Endpoint:     *endpoint,
		EndpointGRPC: *endpointGRPC,
		EndpointHTTP: *endpointHTTP,
		Queue:        *queue,
		Mode:         *mode,
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
