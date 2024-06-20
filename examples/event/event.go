// go run examples/event/event.go

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

	s := &agscheduler.Scheduler{}

	sto := &stores.MemoryStore{}
	err := s.SetStore(sto)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	lis := &agscheduler.Listener{
		Callbacks: []agscheduler.CallbackPkg{
			{
				Callback: jobCallback,
				Event:    agscheduler.EVENT_JOB_ADDED | agscheduler.EVENT_JOB_DELETED,
			},
		},
	}
	err = s.SetListener(lis)
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
	job, _ = s.AddJob(job)

	job, _ = s.PauseJob(job.Id)

	_ = s.DeleteJob(job.Id)

	time.Sleep(1 * time.Second)
}
