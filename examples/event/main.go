// go run examples/event/main.go

package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/examples"
	"github.com/agscheduler/agscheduler/stores"
)

func jobCallback(ep agscheduler.EventPkg) {
	slog.Info(fmt.Sprintf("Event code: `%d`, job `%s`.\n\n", ep.Event, ep.JobId))
}

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

	listener := &agscheduler.Listener{
		Callbacks: []agscheduler.CallbackPkg{
			{
				Callback: jobCallback,
				Event:    agscheduler.EVENT_JOB_ADDED | agscheduler.EVENT_JOB_DELETED,
			},
		},
	}
	err = scheduler.SetListener(listener)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set listener: %s", err))
		os.Exit(1)
	}

	job := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "2s",
		Func:     examples.PrintMsg,
	}
	job, _ = scheduler.AddJob(job)

	job, _ = scheduler.PauseJob(job.Id)

	_ = scheduler.DeleteJob(job.Id)

	time.Sleep(1 * time.Second)
}
