package main

import (
	"log"
	"time"

	"agscheduler"
	"agscheduler/storages"
)

func printMsg(args ...any) {
	log.Println(args...)
}

func main() {
	task := &agscheduler.Task{
		Name:     "Print message",
		Type:     "interval",
		Func:     printMsg,
		Args:     []any{"arg1", "arg2", "arg3"},
		Interval: 1 * time.Second,
	}

	storage := &storages.MemoryStorage{}

	scheduler := &agscheduler.Scheduler{
		Storage: storage,
	}

	scheduler.AddTask(task)
	scheduler.Start()

	task2 := &agscheduler.Task{
		Name:     "Print message2",
		Type:     "cron",
		Func:     printMsg,
		Args:     []any{"arg4", "arg5", "arg6", "arg7"},
		CronExpr: "*/1 * * * *",
	}

	task2Id := scheduler.AddTask(task2)

	time.Sleep(6 * time.Second)
	task2, _ = scheduler.GetTaskById(task2Id)
	task2.Type = "interval"
	task2.Interval = 4 * time.Second
	scheduler.UpdateTask(task2)

	time.Sleep(6 * time.Second)
	scheduler.DeleteTaskById(task2Id)

	select {}
}
