package queues

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

var Ctx = context.Background()

var ExampleQueue = "agscheduler_example_queue"

func RunExample(brk *agscheduler.Broker) {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: examples.PrintMsgSleep},
	)

	s := &agscheduler.Scheduler{}

	sto := &stores.MemoryStore{}
	err := s.SetStore(sto)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(Ctx)
	err = s.SetBroker(ctx, brk)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set broker: %s", err))
		os.Exit(1)
	}

	for i := range 6 {
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

	slog.Info("Sleep 1s......\n\n")
	time.Sleep(1 * time.Second)
	count, _ := brk.CountJobs(ExampleQueue)
	slog.Info(fmt.Sprintf("Number of jobs in this queue: %d\n\n", count))
	slog.Info("Sleep 2s......\n\n")
	time.Sleep(2 * time.Second)
	count, _ = brk.CountJobs(ExampleQueue)
	slog.Info(fmt.Sprintf("Number of jobs in this queue: %d\n\n", count))
	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	_ = s.DeleteAllJobs()

	s.Stop()

	cancel()
	_ = brk.Clear(ExampleQueue)
	_ = sto.Clear()
}
