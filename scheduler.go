package agscheduler

import (
	"fmt"
	"time"

	"github.com/gorhill/cronexpr"
)

func calculateNextRunTime(task *Task) time.Time {
	switch task.Type {
	case "datetime":
		return task.StartAt
	case "interval":
		return time.Now().Add(task.Interval)
	case "cron":
		return cronexpr.MustParse(task.CronExpr).Next(time.Now())
	default:
		panic(fmt.Sprintf("Unknown task type %s", task.Type))
	}
}

func (s *Scheduler) AddTask(task *Task) (id string) {
	if task.NextRunTime.IsZero() {
		task.NextRunTime = calculateNextRunTime(task)
	}

	task.SetId()
	s.Storage.AddTask(task)

	return task.id
}

func (s *Scheduler) GetTaskById(id string) (*Task, error) {
	return s.Storage.GetTaskById(id)
}

func (s *Scheduler) UpdateTask(task *Task) error {
	return s.Storage.UpdateTask(task)
}

func (s *Scheduler) DeleteTaskById(id string) error {
	return s.Storage.DeleteTaskById(id)
}

func (s *Scheduler) run() {
	for {
		select {
		case <-s.QuitChan:
			return
		case <-s.Timer.C:
			now := time.Now()

			for _, task := range s.Storage.GetAllTasks() {
				if task.Status == "paused" {
					continue
				}

				if task.NextRunTime.Before(now) {
					task.LastRunTime = now
					task.NextRunTime = calculateNextRunTime(task)

					go task.Func(task.Args...)
				}
			}

			s.Timer.Reset(time.Second)
		}
	}
}

func (s *Scheduler) Start() {
	s.Timer = time.NewTimer(0)
	s.QuitChan = make(chan struct{})

	go s.run()
}
