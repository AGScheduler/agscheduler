package main

import (
	"log"
	"time"

	"agscheduler"
	"agscheduler/stores"
)

func printMsg(args ...any) {
	log.Println(args...)
}

func main() {
	job := &agscheduler.Job{
		Name:     "Print message",
		Type:     "interval",
		Func:     printMsg,
		Args:     []any{"arg1", "arg2", "arg3"},
		Interval: 1 * time.Second,
	}

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
    scheduler.SetStore(store)

	scheduler.AddJob(job)
	scheduler.Start()

	job2 := &agscheduler.Job{
		Name:     "Print message2",
		Type:     "cron",
		Func:     printMsg,
		Args:     []any{"arg4", "arg5", "arg6", "arg7"},
		CronExpr: "*/1 * * * *",
	}

	job2Id := scheduler.AddJob(job2)

	time.Sleep(6 * time.Second)
	job2, _ = scheduler.GetJobById(job2Id)
	job2.Type = "interval"
	job2.Interval = 4 * time.Second
	scheduler.UpdateJob(job2)

	time.Sleep(6 * time.Second)
	scheduler.DeleteJobById(job2Id)

	select {}
}
