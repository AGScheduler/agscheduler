// go run examples/cluster/cluster_worker.go

package main

import (
	"flag"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/examples"
	"github.com/kwkwc/agscheduler/services"
	"github.com/kwkwc/agscheduler/stores"
)

var mainEndpoint = flag.String("me", "127.0.0.1:36364", "Cluster Main endpoint")
var endpoint = flag.String("e", "127.0.0.1:36366", "Cluster Worker endpoint")
var schedulerEndpoint = flag.String("se", "127.0.0.1:36365", "Cluster Worker Scheduler endpoint")
var schedulerQueue = flag.String("sq", "worker", "Cluster Worker Scheduler queue")

func main() {
	agscheduler.RegisterFuncs(examples.PrintMsg)

	flag.Parse()

	store := &stores.MemoryStore{}

	cw := &agscheduler.ClusterWorker{
		MainEndpoint:      *mainEndpoint,
		Endpoint:          *endpoint,
		SchedulerEndpoint: *schedulerEndpoint,
		SchedulerQueue:    *schedulerQueue,
	}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)
	scheduler.SetClusterWorker(cw)

	rservice := &services.SchedulerRPCService{
		Scheduler: scheduler,
		Address:   cw.SchedulerEndpoint,
		Queue:     cw.SchedulerQueue,
	}

	crservice := services.ClusterRPCService{
		Srs: rservice,
		Cw:  cw,
	}
	crservice.Start()

	cw.Register()

	select {}
}
