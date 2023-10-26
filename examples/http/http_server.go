// go run examples/http/http_server.go

package main

import (
	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/examples"
	"github.com/kwkwc/agscheduler/services"
	"github.com/kwkwc/agscheduler/stores"
)

func main() {
	agscheduler.RegisterFuncs(examples.PrintMsg)

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	hservice := services.SchedulerHTTPService{
		Scheduler: scheduler,
		Address:   "127.0.0.1:63636",
	}
	hservice.Start()

	select {}
}
