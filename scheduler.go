package agscheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/gorhill/cronexpr"
)

type Scheduler struct {
	store    Store
	timer    *time.Timer
	quitChan chan struct{}
}

func (s *Scheduler) SetStore(sto Store) {
	s.store = sto
}

func CalcNextRunTime(j Job) time.Time {
	if j.Status == "paused" {
		nextRunTime, _ := time.Parse("2006-01-02 15:04:05", "9999-09-09 09:09:09")
		return nextRunTime
	}
	switch j.Type {
	case "datetime":
		return j.StartAt
	case "interval":
		return time.Now().Add(j.Interval)
	case "cron":
		return cronexpr.MustParse(j.CronExpr).Next(time.Now())
	default:
		panic(fmt.Sprintf("Unknown job type %s", j.Type))
	}
}

func (s *Scheduler) AddJob(j Job) (id string) {
	j.SetId()
	j.Status = "running"

	if j.NextRunTime.IsZero() {
		j.NextRunTime = CalcNextRunTime(j)
	}

	s.store.AddJob(j)

	return j.id
}

func (s *Scheduler) GetJob(id string) (Job, error) {
	return s.store.GetJob(id)
}

func (s *Scheduler) UpdateJob(j Job) error {
	err := s.store.UpdateJob(j)
	s.wakeup()
	return err
}

func (s *Scheduler) DeleteJob(id string) error {
	return s.store.DeleteJob(id)
}

func (s *Scheduler) PauseJob(id string) error {
	j, err := s.GetJob(id)
	if err != nil {
		return err
	}

	j.Status = "paused"

	err = s.UpdateJob(j)
	if err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) ResumeJob(id string) error {
	j, err := s.GetJob(id)
	if err != nil {
		return err
	}

	j.Status = "running"

	err = s.UpdateJob(j)
	if err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) run() {
	for {
		select {
		case <-s.quitChan:
			return
		case <-s.timer.C:
			now := time.Now()

			for _, j := range s.store.GetAllJobs() {
				if j.Status == "paused" {
					continue
				}

				if j.NextRunTime.Before(now) {
					j.NextRunTime = CalcNextRunTime(j)

					go j.Func(j)

					j.LastRunTime = now

					if j.Type == "datetime" {
						s.DeleteJob(j.id)
					} else {
						err := s.UpdateJob(j)
						if err != nil {
							log.Println("Scheduler run error:", err)
							continue
						}
					}
				}
			}

			minNextRunTime := s.store.GetNextRunTime()
			now = time.Now()
			nextWakeupInterval := minNextRunTime.Sub(now)
			if nextWakeupInterval < 0 {
				nextWakeupInterval = time.Second
			}
			s.timer.Reset(nextWakeupInterval)
		}
	}
}

func (s *Scheduler) Start() {
	s.timer = time.NewTimer(0)
	s.quitChan = make(chan struct{})

	go s.run()
}

func (s *Scheduler) Stop() {
	s.quitChan <- struct{}{}
}

func (s *Scheduler) wakeup() {
	s.timer.Reset(0)
}
