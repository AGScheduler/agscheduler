// go run examples/http/http_server.go

package main

import (
	"fmt"
	"log/slog"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/services"
	"github.com/kwkwc/agscheduler/stores"
)

func printMsg(j agscheduler.Job) {
	slog.Info(fmt.Sprintf("Run job `%s` %s\n\n", j.FullName(), j.Args))
}

func main() {
	agscheduler.RegisterFuncs(printMsg)

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
