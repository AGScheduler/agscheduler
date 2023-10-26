// go run examples/cluster/cluster_main.go

package main

import (
	"flag"

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

	store := &stores.MemoryStore{}

	cm := &agscheduler.ClusterMain{
		Endpoint:          *endpoint,
		SchedulerEndpoint: *schedulerEndpoint,
		SchedulerQueue:    *schedulerQueue,
	}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)
	scheduler.SetClusterMain(cm)

	rservice := &services.SchedulerRPCService{
		Scheduler: scheduler,
		Address:   cm.SchedulerEndpoint,
		Queue:     cm.SchedulerQueue,
	}

	crservice := services.ClusterRPCService{
		Srs: rservice,
		Cm:  cm,
	}
	crservice.Start()

	select {}
}
