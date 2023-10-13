package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"
)

func printMsg(j agscheduler.Job) {
	slog.Info(fmt.Sprintf("Run %s %s\n", j.Name, j.Args))
}

func main() {
	agscheduler.RegisterFuncs(printMsg)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0,
		Password: "",
	})
	store := &stores.RedisStore{RDB: rdb}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	job1 := agscheduler.Job{
		Name:     "Job1",
		Type:     agscheduler.TYPE_INTERVAL,
		Timezone: "Asia/Shanghai",
		Func:     printMsg,
		Args:     []any{"arg1", "arg2", "arg3"},
		Interval: 2 * time.Second,
	}
	job1Id := scheduler.AddJob(job1)
	job1, _ = scheduler.GetJob(job1Id)
	slog.Info(fmt.Sprintf("Scheduler add %s %s.\n\n", job1.Name, job1))

	job2 := agscheduler.Job{
		Name:     "Job2",
		Type:     agscheduler.TYPE_CRON,
		Func:     printMsg,
		Args:     []any{"arg4", "arg5", "arg6", "arg7"},
		CronExpr: "*/1 * * * *",
	}
	job2Id := scheduler.AddJob(job2)
	job2, _ = scheduler.GetJob(job2Id)
	slog.Info(fmt.Sprintf("Scheduler add %s %s.\n\n", job2.Name, job2))

	scheduler.Start()
	slog.Info("Scheduler start.\n\n")

	timezone, _ := time.LoadLocation("America/New_York")
	startAt, _ := time.ParseInLocation("2006-01-02 15:04:05", "2023-09-22 07:30:08", timezone)
	job3 := agscheduler.Job{
		Name:     "Job3",
		Type:     agscheduler.TYPE_DATETIME,
		Timezone: timezone.String(),
		Func:     printMsg,
		Args:     []any{"arg8", "arg9"},
		StartAt:  startAt,
	}
	job3Id := scheduler.AddJob(job3)
	job3, _ = scheduler.GetJob(job3Id)
	slog.Info(fmt.Sprintf("Scheduler add %s %s.\n\n", job3.Name, job3))

	slog.Info("Sleep 10s......\n\n")
	time.Sleep(10 * time.Second)

	job2, _ = scheduler.GetJob(job2Id)
	job2.Type = agscheduler.TYPE_INTERVAL
	job2.Interval = 4 * time.Second
	scheduler.UpdateJob(job2)
	job2, _ = scheduler.GetJob(job2Id)
	slog.Info(fmt.Sprintf("Scheduler update %s %s.\n\n", job2.Name, job2))

	slog.Info("Sleep 8s......")
	time.Sleep(8 * time.Second)

	scheduler.PauseJob(job1Id)
	slog.Info(fmt.Sprintf("Scheduler pause %s.\n\n", job1.Name))

	slog.Info("Sleep 6s......\n\n")
	time.Sleep(6 * time.Second)

	scheduler.ResumeJob(job1Id)
	slog.Info(fmt.Sprintf("Scheduler resume %s.\n\n", job1.Name))

	scheduler.DeleteJob(job2Id)
	slog.Info(fmt.Sprintf("Scheduler delete %s.\n\n", job2.Name))

	slog.Info("Sleep 6s......\n\n")
	time.Sleep(6 * time.Second)

	scheduler.Stop()
	slog.Info("Scheduler stop.\n\n")

	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	scheduler.Start()
	slog.Info("Scheduler start.\n\n")

	slog.Info("Sleep 4s......\n\n")
	time.Sleep(4 * time.Second)

	scheduler.DeleteAllJobs()
	slog.Info("Scheduler delete all jobs.\n\n")

	select {}
}
