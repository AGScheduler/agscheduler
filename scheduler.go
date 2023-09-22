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
	if job.NextRunTime.IsZero() {
		job.NextRunTime = calcNextRunTime(job)
	}

	job.SetId()
	s.store.AddJob(job)

	return job.id
}

func (s *Scheduler) GetJobById(id string) (*Job, error) {
	return s.store.GetJobById(id)
}

func (s *Scheduler) UpdateJob(job *Job) error {
	return s.store.UpdateJob(job)
}

func (s *Scheduler) DeleteJobById(id string) error {
	return s.store.DeleteJobById(id)
}

func (s *Scheduler) run() {
	for {
		select {
		case <-s.quitChan:
			return
		case <-s.timer.C:
			now := time.Now()

			for _, Job := range s.store.GetAllJobs() {
				if Job.Status == "paused" {
					continue
				}

				if Job.NextRunTime.Before(now) {
					Job.LastRunTime = now
					Job.NextRunTime = calcNextRunTime(Job)

					go Job.Func(Job.Args...)
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
