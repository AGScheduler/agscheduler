package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/examples"
	"github.com/agscheduler/agscheduler/stores"
)

var ctx = context.Background()

func runExample(rec *agscheduler.Recorder) {
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
	err = s.SetRecorder(rec)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set backend: %s", err))
		os.Exit(1)
	}

	job := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "2s",
		Func:     examples.PrintMsg,
	}
	job, _ = s.AddJob(job)
	slog.Info(fmt.Sprintf("%s.\n\n", job))

	job2 := agscheduler.Job{
		Name:    "Job",
		Type:    agscheduler.JOB_TYPE_DATETIME,
		StartAt: "2023-09-22 07:30:08",
		Func:    examples.PrintMsg,
	}
	job2, _ = s.AddJob(job2)
	slog.Info(fmt.Sprintf("%s.\n\n", job2))

	s.Start()

	slog.Info("Sleep 5s......\n\n")
	time.Sleep(5 * time.Second)

	records, _, _ := rec.GetRecords(job.Id, 1, 10)
	slog.Info(fmt.Sprintf("Scheduler recorder get records %v.\n\n", records))

	records, _, _ = rec.GetAllRecords(1, 10)
	slog.Info(fmt.Sprintf("Scheduler recorder get all records %v.\n\n", records))

	rec.DeleteRecords(job.Id)
	rec.DeleteAllRecords()

	s.DeleteAllJobs()

	s.Stop()

	sto.Clear()
	rec.Clear()
}
