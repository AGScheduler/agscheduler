package agscheduler

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"time"

	"github.com/gorhill/cronexpr"
)

type Scheduler struct {
	store    Store
	ticker   *time.Timer
	quitChan chan struct{}
}

func (s *Scheduler) SetStore(sto Store) {
	s.store = sto
	s.store.Init()
}

func CalcNextRunTime(j Job) time.Time {
	timezone, _ := time.LoadLocation(j.Timezone)
	if j.Status == "paused" {
		nextRunTime, _ := time.ParseInLocation("2006-01-02 15:04:05", "9999-09-09 09:09:09", timezone)
		return nextRunTime
	}
	switch j.Type {
	case "datetime":
		return j.StartAt.In(timezone)
	case "interval":
		return time.Now().In(timezone).Add(j.Interval)
	case "cron":
		return cronexpr.MustParse(j.CronExpr).Next(time.Now().In(timezone))
	default:
		panic(fmt.Sprintf("Unknown job type %s", j.Type))
	}
}

func (s *Scheduler) AddJob(j Job) (id string) {
	j.SetId()
	j.Status = "running"

	if j.Timezone == "" {
		j.Timezone = "UTC"
	}

	if j.NextRunTime.IsZero() {
		j.NextRunTime = CalcNextRunTime(j)
	}

	j.FuncName = runtime.FuncForPC(reflect.ValueOf(j.Func).Pointer()).Name()

	s.store.AddJob(j)

	return j.Id
}

func (s *Scheduler) GetJob(id string) (Job, error) {
	return s.store.GetJob(id)
}

func (s *Scheduler) UpdateJob(j Job) error {
	lastNextWakeupInterval := s.getNextWakeupInterval()

	err := s.store.UpdateJob(j)

	nextWakeupInterval := s.getNextWakeupInterval()
	if nextWakeupInterval < lastNextWakeupInterval {
		s.wakeup()
	}

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
		case <-s.ticker.C:
			now := time.Now()

			jobs, _ := s.store.GetAllJobs()
			for _, j := range jobs {
				if j.Status == "paused" {
					continue
				}

				timezone, _ := time.LoadLocation(j.Timezone)
				now := now.In(timezone)

				if j.NextRunTime.Before(now) {
					j.NextRunTime = CalcNextRunTime(j)

					f := reflect.ValueOf(funcs[j.FuncName])
					go f.Call([]reflect.Value{reflect.ValueOf(j)})

					j.LastRunTime = now

					if j.Type == "datetime" {
						s.DeleteJob(j.Id)
					} else {
						err := s.UpdateJob(j)
						if err != nil {
							log.Println("Scheduler run error:", err)
							continue
						}
					}
				}
			}

			nextWakeupInterval := s.getNextWakeupInterval()
			s.ticker.Reset(nextWakeupInterval)
		}
	}
}

func (s *Scheduler) Start() {
	s.ticker = time.NewTimer(0)
	s.quitChan = make(chan struct{})

	go s.run()
}

func (s *Scheduler) Stop() {
	s.quitChan <- struct{}{}
}

func (s *Scheduler) getNextWakeupInterval() time.Duration {
	minNextRunTime, _ := s.store.GetNextRunTime()
	now := time.Now().In(minNextRunTime.Location())
	nextWakeupInterval := minNextRunTime.Sub(now)
	if nextWakeupInterval < 0 {
		nextWakeupInterval = time.Second
	}
	return nextWakeupInterval
}

func (s *Scheduler) wakeup() {
	s.ticker.Reset(0)
}
