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

func calcNextRunTime(job *Job) time.Time {
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
		job.NextRunTime = calcNextRunTime(job)
	}

	s.store.AddJob(job)

	return job.id
}

func (s *Scheduler) GetJob(id string) (*Job, error) {
	return s.store.GetJob(id)
}

func (s *Scheduler) UpdateJob(job *Job) error {
	return s.store.UpdateJob(job)
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
					j.NextRunTime = calcNextRunTime(j)

					t := *j
					t.Func = nil
					go j.Func(t)

					if j.Type == "datetime" {
						s.DeleteJob(j.id)
					}
				}
			}

			s.timer.Reset(time.Second)
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
