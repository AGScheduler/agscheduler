package stores

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/examples"
)

var Ctx = context.Background()

func RunExample(sto agscheduler.Store) {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: examples.PrintMsg},
	)

	s := &agscheduler.Scheduler{}

	err := s.SetStore(sto)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	job1 := agscheduler.Job{
		Name:     "Job1",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "2s",
		Timezone: "UTC",
		Func:     examples.PrintMsg,
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	job1, _ = s.AddJob(job1)
	slog.Info(fmt.Sprintf("%s.\n\n", job1))

	job2 := agscheduler.Job{
		Name:     "Job2",
		Type:     agscheduler.JOB_TYPE_CRON,
		CronExpr: "*/1 * * * *",
		Timezone: "Asia/Shanghai",
		FuncName: "github.com/agscheduler/agscheduler/examples.PrintMsg",
		Args:     map[string]any{"arg4": "4", "arg5": "5", "arg6": "6", "arg7": "7"},
	}
	job2, _ = s.AddJob(job2)
	slog.Info(fmt.Sprintf("%s.\n\n", job2))

	s.Start()

	job3 := agscheduler.Job{
		Name:     "Job3",
		Type:     agscheduler.JOB_TYPE_DATETIME,
		StartAt:  "2023-09-22 07:30:08",
		Timezone: "America/New_York",
		Func:     examples.PrintMsg,
		Args:     map[string]any{"arg8": "8", "arg9": "9"},
	}
	job3, _ = s.AddJob(job3)
	slog.Info(fmt.Sprintf("%s.\n\n", job3))

	jobs, _ := s.GetAllJobs()
	slog.Info(fmt.Sprintf("Scheduler get all jobs %s.\n\n", jobs))

	slog.Info("Sleep 5s......\n\n")
	time.Sleep(5 * time.Second)

	job1, _ = s.GetJob(job1.Id)
	slog.Info(fmt.Sprintf("Scheduler get job `%s` %s.\n\n", job1.FullName(), job1))

	job2.Type = agscheduler.JOB_TYPE_INTERVAL
	job2.Interval = "3s"
	job2, _ = s.UpdateJob(job2)
	slog.Info(fmt.Sprintf("Scheduler update job `%s` %s.\n\n", job2.FullName(), job2))

	slog.Info("Sleep 4s......")
	time.Sleep(4 * time.Second)

	job1, _ = s.PauseJob(job1.Id)

	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	job1, _ = s.ResumeJob(job1.Id)

	_ = s.DeleteJob(job2.Id)

	slog.Info("Sleep 4s......\n\n")
	time.Sleep(4 * time.Second)

	s.Stop()

	_ = s.RunJob(job1)

	_ = s.ScheduleJob(job1)

	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	s.Start()

	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	_ = s.DeleteAllJobs()

	s.Stop()

	_ = sto.Clear()
}
