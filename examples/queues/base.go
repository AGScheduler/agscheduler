package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/examples"
	"github.com/agscheduler/agscheduler/stores"
)

var ctx = context.Background()

var exampleQueue = "agscheduler_example_queue"

func runExample(brk *agscheduler.Broker) {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: examples.PrintMsgSleep},
	)

	sto := &stores.MemoryStore{}
	s := &agscheduler.Scheduler{}
	err := s.SetStore(sto)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(ctx)
	err = s.SetBroker(ctx, brk)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set broker: %s", err))
		os.Exit(1)
	}

	for i := range 5 {
		job := agscheduler.Job{
			Name:    "Job" + strconv.Itoa(i+1),
			Type:    agscheduler.JOB_TYPE_DATETIME,
			StartAt: "2023-09-22 07:30:08",
			Func:    examples.PrintMsgSleep,
		}
		job, _ = s.AddJob(job)
		slog.Info(fmt.Sprintf("%s.\n\n", job))
	}

	s.Start()

	slog.Info("Sleep 6s......\n\n")
	time.Sleep(6 * time.Second)

	s.DeleteAllJobs()

	s.Stop()

	cancel()
	brk.Queues[exampleQueue].Clear()
	sto.Clear()
}
