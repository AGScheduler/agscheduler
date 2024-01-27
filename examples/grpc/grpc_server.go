// go run examples/grpc/grpc_server.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/examples"
	"github.com/kwkwc/agscheduler/services"
	"github.com/kwkwc/agscheduler/stores"
)

func main() {
	agscheduler.RegisterFuncs(examples.PrintMsg)

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	grservice := services.GRPCService{
		Scheduler: scheduler,
		Address:   "127.0.0.1:36360",
	}
	err = grservice.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start service: %s", err))
		os.Exit(1)
	}

	select {}
}
