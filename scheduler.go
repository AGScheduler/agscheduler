package agscheduler

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
)

type Scheduler struct {
	store     Store
	timer     *time.Timer
	quitChan  chan struct{}
	isRunning bool
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
	switch strings.ToLower(j.Type) {
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

	s.Start()

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

func (s *Scheduler) DeleteAllJobs() error {
	return s.store.DeleteAllJobs()
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

			jobs, _ := s.store.GetAllJobs()
			if len(jobs) == 0 {
				s.Stop()
				return
			}

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
			s.timer.Reset(nextWakeupInterval)
		}
	}
}

func (s *Scheduler) Start() {
	if s.isRunning {
		return
	}

	s.timer = time.NewTimer(0)
	s.quitChan = make(chan struct{})
	s.isRunning = true

	go s.run()
}

func (s *Scheduler) Stop() {
	if !s.isRunning {
		return
	}

	s.isRunning = false
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
	s.timer.Reset(0)
}
