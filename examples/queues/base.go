package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/examples"
)

func runExample(s *agscheduler.Scheduler) {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: examples.PrintMsgSleep},
	)

	for i := range 5 {
		job := agscheduler.Job{
			Name:    "Job" + strconv.Itoa(i+1),
			Type:    agscheduler.TYPE_DATETIME,
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
}
