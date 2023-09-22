package agscheduler

import (
	"fmt"
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

func CalcNextRunTime(job *Job) time.Time {
	if job.Status == "paused" {
		nextRunTime, _ := time.Parse("2006-01-02 15:04:05", "9999-09-09 09:09:09")
		return nextRunTime
	}
	switch job.Type {
	case "datetime":
		return job.StartAt
	case "interval":
		return time.Now().Add(job.Interval)
	case "cron":
		return cronexpr.MustParse(job.CronExpr).Next(time.Now())
	default:
		panic(fmt.Sprintf("Unknown job type %s", job.Type))
	}
}

func (s *Scheduler) AddJob(job *Job) (id string) {
	job.SetId()
	job.Status = "running"

	if job.NextRunTime.IsZero() {
		job.NextRunTime = CalcNextRunTime(job)
	}

	s.store.AddJob(job)

	return job.id
}

func (s *Scheduler) GetJob(id string) (*Job, error) {
	return s.store.GetJob(id)
}

func (s *Scheduler) UpdateJob(job *Job) error {
	err := s.store.UpdateJob(job)
	s.wakeup()
	return err
}

func (s *Scheduler) DeleteJob(id string) error {
	return s.store.DeleteJob(id)
}

func (s *Scheduler) PauseJob(id string) error {
	job, err := s.GetJob(id)
	if err != nil {
		return err
	}
	job.Status = "paused"
	err = s.UpdateJob(job)
	if err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) ResumeJob(id string) error {
	job, err := s.GetJob(id)
	if err != nil {
		return err
	}
	job.Status = "running"
	err = s.UpdateJob(job)
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
					j.LastRunTime = now
					j.NextRunTime = CalcNextRunTime(j)

					t := *j
					t.Func = nil
					go j.Func(t)

					if j.Type == "datetime" {
						s.DeleteJob(j.id)
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
