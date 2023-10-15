package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/kwkwc/agscheduler"
)

func printMsg(j agscheduler.Job) {
	slog.Info(fmt.Sprintf("Run %s %s\n", j.Name, j.Args))
}

func runExample(s *agscheduler.Scheduler) {
	agscheduler.RegisterFuncs(printMsg)

	job1 := agscheduler.Job{
		Name:     "Job1",
		Type:     agscheduler.TYPE_INTERVAL,
		Timezone: "Asia/Shanghai",
		Func:     printMsg,
		Args:     []any{"arg1", "arg2", "arg3"},
		Interval: 2 * time.Second,
	}
	job1Id, _ := s.AddJob(job1)
	job1, _ = s.GetJob(job1Id)
	slog.Info(fmt.Sprintf("Scheduler add %s %s.\n\n", job1.Name, job1))

	job2 := agscheduler.Job{
		Name:     "Job2",
		Type:     agscheduler.TYPE_CRON,
		Func:     printMsg,
		Args:     []any{"arg4", "arg5", "arg6", "arg7"},
		CronExpr: "*/1 * * * *",
	}
	job2Id, _ := s.AddJob(job2)
	job2, _ = s.GetJob(job2Id)
	slog.Info(fmt.Sprintf("Scheduler add %s %s.\n\n", job2.Name, job2))

	s.Start()
	slog.Info("Scheduler start.\n\n")

	timezone, _ := time.LoadLocation("America/New_York")
	startAt, _ := time.ParseInLocation(time.DateTime, "2023-09-22 07:30:08", timezone)
	job3 := agscheduler.Job{
		Name:     "Job3",
		Type:     agscheduler.TYPE_DATETIME,
		Timezone: timezone.String(),
		Func:     printMsg,
		Args:     []any{"arg8", "arg9"},
		StartAt:  startAt,
	}
	job3Id, _ := s.AddJob(job3)
	job3, _ = s.GetJob(job3Id)
	slog.Info(fmt.Sprintf("Scheduler add %s %s.\n\n", job3.Name, job3))

	slog.Info("Sleep 10s......\n\n")
	time.Sleep(10 * time.Second)

	job2, _ = s.GetJob(job2Id)
	job2.Type = agscheduler.TYPE_INTERVAL
	job2.Interval = 4 * time.Second
	s.UpdateJob(job2)
	job2, _ = s.GetJob(job2Id)
	slog.Info(fmt.Sprintf("Scheduler update %s %s.\n\n", job2.Name, job2))

	slog.Info("Sleep 8s......")
	time.Sleep(8 * time.Second)

	s.PauseJob(job1Id)
	slog.Info(fmt.Sprintf("Scheduler pause %s.\n\n", job1.Name))

	slog.Info("Sleep 6s......\n\n")
	time.Sleep(6 * time.Second)

	s.ResumeJob(job1Id)
	slog.Info(fmt.Sprintf("Scheduler resume %s.\n\n", job1.Name))

	s.DeleteJob(job2Id)
	slog.Info(fmt.Sprintf("Scheduler delete %s.\n\n", job2.Name))

	slog.Info("Sleep 6s......\n\n")
	time.Sleep(6 * time.Second)

	s.Stop()
	slog.Info("Scheduler stop.\n\n")

	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	s.Start()
	slog.Info("Scheduler start.\n\n")

	slog.Info("Sleep 4s......\n\n")
	time.Sleep(4 * time.Second)

	s.DeleteAllJobs()
	slog.Info("Scheduler delete all jobs.\n\n")
}
