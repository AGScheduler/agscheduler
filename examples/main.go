package main

import (
	"log"
	"time"

	"agscheduler"
	"agscheduler/stores"
)

func printMsg(j agscheduler.Job) {
	log.Printf("Run %s %s\n", j.Name, j.Args)
}

func main() {
	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	job1 := agscheduler.Job{
		Name:     "Job1",
		Type:     "interval",
		Func:     printMsg,
		Args:     []any{"arg1", "arg2", "arg3"},
		Interval: 2 * time.Second,
	}
	job1Id := scheduler.AddJob(job1)
	job1, _ = scheduler.GetJob(job1Id)
	log.Printf("Scheduler add %s %s.\n\n", job1.Name, job1)

	job2 := agscheduler.Job{
		Name:     "Job2",
		Type:     "cron",
		Func:     printMsg,
		Args:     []any{"arg4", "arg5", "arg6", "arg7"},
		CronExpr: "*/1 * * * *",
	}
	job2Id := scheduler.AddJob(job2)
	job2, _ = scheduler.GetJob(job2Id)
	log.Printf("Scheduler add %s %s.\n\n", job2.Name, job2)

	scheduler.Start()
	log.Print("Scheduler Start.\n\n")

	startAt, _ := time.Parse("2006-01-02 15:04:05", "2023-09-22 07:30:08")
	job3 := agscheduler.Job{
		Name:    "Job3",
		Type:    "datetime",
		Func:    printMsg,
		Args:    []any{"arg8", "arg9"},
		StartAt: startAt,
	}
	job3Id := scheduler.AddJob(job3)
	job3, _ = scheduler.GetJob(job3Id)
	log.Printf("Scheduler add %s %s.\n\n", job3.Name, job3)

	log.Print("Sleep 10s......\n\n")
	time.Sleep(10 * time.Second)

	job2, _ = scheduler.GetJob(job2Id)
	job2.Type = "interval"
	job2.Interval = 4 * time.Second
	scheduler.UpdateJob(job2)
	log.Printf("Scheduler update %s %s.\n\n", job2.Name, job2)

	log.Println("Sleep 8s......")
	time.Sleep(8 * time.Second)

	scheduler.PauseJob(job1Id)
	log.Printf("Scheduler pause %s.\n\n", job1.Name)

	log.Print("Sleep 6s......\n\n")
	time.Sleep(6 * time.Second)

	scheduler.ResumeJob(job1Id)
	log.Printf("Scheduler resume %s.\n\n", job1.Name)

	scheduler.DeleteJob(job2Id)
	log.Printf("Scheduler delete %s.\n\n", job2.Name)

	log.Print("Sleep 6s......\n\n")
	time.Sleep(6 * time.Second)

	scheduler.Stop()
	log.Print("Scheduler Stop.\n\n")

	log.Print("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	scheduler.Start()
	log.Print("Scheduler Start.\n\n")

	select {}
}
