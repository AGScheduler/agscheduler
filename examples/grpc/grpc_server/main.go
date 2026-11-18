// go run examples/grpc/grpc_server/main.go

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/examples"
	"github.com/agscheduler/agscheduler/services"
	"github.com/agscheduler/agscheduler/stores"
)

func main() {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: examples.PrintMsg},
	)

	scheduler := &agscheduler.Scheduler{}

	store := &stores.MemoryStore{}
	err := scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	grservice := services.GRPCService{
		Scheduler: scheduler,
		Address:   "127.0.0.1:36360",
		// PasswordSha2: "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918",
	}
	err = grservice.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start service: %s", err))
		os.Exit(1)
	}

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	<-signals
	slog.Info("Shutdown signal received, waiting for running jobs.")
	go func() {
		<-signals
		slog.Warn("Second shutdown signal received, forcing exit.")
		os.Exit(1)
	}()

	grservice.Stop()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), agscheduler.DefaultShutdownTimeout)
	defer cancelShutdown()
	if err := scheduler.Shutdown(shutdownCtx); err != nil {
		slog.Warn(fmt.Sprintf("Scheduler graceful shutdown failed: %s", err))
	}
}
