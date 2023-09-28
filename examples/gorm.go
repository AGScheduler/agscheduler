package main

import (
	"log"
	"time"

	// "gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"
)

func printMsg(j agscheduler.Job) {
	log.Printf("Run %s %s\n", j.Name, j.Args)
}

func main() {
	agscheduler.RegisterFuncs(printMsg)

	// dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=UTC"
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	db, err := gorm.Open(sqlite.Open("agscheduler.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}
	store := &stores.GORMStore{DB: db}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	job1 := agscheduler.Job{
		Name:     "Job1",
		Type:     "interval",
		Timezone: "Asia/Shanghai",
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
	log.Print("Scheduler start.\n\n")

	timezone, _ := time.LoadLocation("America/New_York")
	startAt, _ := time.ParseInLocation("2006-01-02 15:04:05", "2023-09-22 07:30:08", timezone)
	job3 := agscheduler.Job{
		Name:    "Job3",
		Type:    "datetime",
		Timezone: timezone.String(),
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
	job2, _ = scheduler.GetJob(job2Id)
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
	log.Print("Scheduler stop.\n\n")

	log.Print("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	scheduler.Start()
	log.Print("Scheduler start.\n\n")

	select {}
}
