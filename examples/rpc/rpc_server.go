// go run examples/rpc/rpc_server.go

package main

import (
	"fmt"
	"log/slog"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/services"
	"github.com/kwkwc/agscheduler/stores"
)

func printMsg(j agscheduler.Job) {
	slog.Info(fmt.Sprintf("Run job `%s` %s\n", j.FullName(), j.Args))
}

func main() {
	agscheduler.RegisterFuncs(printMsg)

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	service := services.SchedulerRPCService{Scheduler: scheduler}
	service.Start("")

	select {}
}
