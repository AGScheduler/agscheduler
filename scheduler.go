package agscheduler

import (
	"fmt"
	"log/slog"
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

func (s *Scheduler) SetStore(sto Store) error {
	s.store = sto
	if err := s.store.Init(); err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) Store() Store {
	return s.store
}

func CalcNextRunTime(j Job) (time.Time, error) {
	timezone, err := time.LoadLocation(j.Timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("job `%s` timezone `%s` error: %s", j.Id, j.Timezone, err)
	}
	if j.Status == STATUS_PAUSED {
		nextRunTimeMax, _ := time.ParseInLocation(time.DateTime, "9999-09-09 09:09:09", timezone)
		return time.Unix(nextRunTimeMax.Unix(), 0), nil
	}

	var nextRunTime time.Time
	switch strings.ToLower(j.Type) {
	case TYPE_DATETIME:
		nextRunTime = j.StartAt.In(timezone)
	case TYPE_INTERVAL:
		nextRunTime = time.Now().In(timezone).Add(j.Interval)
	case TYPE_CRON:
		nextRunTime = cronexpr.MustParse(j.CronExpr).Next(time.Now().In(timezone))
	default:
		return time.Time{}, fmt.Errorf("unknown job type %s", j.Type)
	}

	return time.Unix(nextRunTime.Unix(), 0), nil
}

func (s *Scheduler) AddJob(j Job) (id string, err error) {
	j.SetId()
	j.Status = STATUS_RUNNING

	if j.Timezone == "" {
		j.Timezone = "UTC"
	}

	if j.NextRunTime.IsZero() {
		nextRunTime, err := CalcNextRunTime(j)
		if err != nil {
			return id, err
		}
		j.NextRunTime = nextRunTime
	}

	j.FuncName = runtime.FuncForPC(reflect.ValueOf(j.Func).Pointer()).Name()

	if err := s.store.AddJob(j); err != nil {
		return id, err
	}

	s.Start()

	return j.Id, nil
}

func (s *Scheduler) GetJob(id string) (Job, error) {
	return s.store.GetJob(id)
}

func (s *Scheduler) GetAllJobs() ([]Job, error) {
	return s.store.GetAllJobs()
}

func (s *Scheduler) UpdateJob(j Job) error {
	if _, err := s.GetJob(j.Id); err != nil {
		return err
	}

	lastNextWakeupInterval := s.getNextWakeupInterval()

	err := s.store.UpdateJob(j)

	nextWakeupInterval := s.getNextWakeupInterval()
	if nextWakeupInterval < lastNextWakeupInterval {
		s.wakeup()
	}

	return err
}

func (s *Scheduler) DeleteJob(id string) error {
	if _, err := s.GetJob(id); err != nil {
		return err
	}

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

	j.Status = STATUS_PAUSED

	if err := s.UpdateJob(j); err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) ResumeJob(id string) error {
	j, err := s.GetJob(id)
	if err != nil {
		return err
	}

	j.Status = STATUS_RUNNING

	if err := s.UpdateJob(j); err != nil {
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

			jobs, err := s.GetAllJobs()
			if err != nil {
				slog.Error(fmt.Sprintf("Get all jobs error: %s\n", err))
				continue
			}
			if len(jobs) == 0 {
				s.Stop()
				return
			}

			for _, j := range jobs {
				if j.Status == STATUS_PAUSED {
					continue
				}

				timezone, err := time.LoadLocation(j.Timezone)
				if err != nil {
					slog.Error(fmt.Sprintf("Job `%s` timezone `%s` error: %s\n", j.Id, j.Timezone, err))
					continue
				}
				now := now.In(timezone)

				if j.NextRunTime.Before(now) {
					nextRunTime, err := CalcNextRunTime(j)
					if err != nil {
						slog.Error(fmt.Sprintf("Calc next run time error: %s\n", err))
						continue
					}
					j.NextRunTime = nextRunTime

					f := reflect.ValueOf(funcs[j.FuncName])
					if f.IsNil() {
						slog.Warn(fmt.Sprintf("Job `%s` Func is nil\n", j.Id))
					} else {
						go f.Call([]reflect.Value{reflect.ValueOf(j)})
					}

					j.LastRunTime = time.Unix(now.Unix(), 0)

					if j.Type == TYPE_DATETIME {
						err := s.DeleteJob(j.Id)
						if err != nil {
							slog.Error(fmt.Sprintf("Delete job `%s` error: %s\n", j.Id, err))
							continue
						}
					} else {
						err := s.UpdateJob(j)
						if err != nil {
							slog.Error(fmt.Sprintf("Update job `%s` error: %s\n", j.Id, err))
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
	nextRunTimeMin, err := s.store.GetNextRunTime()
	if err != nil {
		slog.Error(fmt.Sprintf("Get next wakeup interval error: %s\n", err))
		nextRunTimeMin = time.Now().UTC().Add(1 * time.Second)
	}

	now := time.Now().In(nextRunTimeMin.Location())
	nextWakeupInterval := nextRunTimeMin.Sub(now)
	if nextWakeupInterval < 0 {
		nextWakeupInterval = time.Second
	}

	return nextWakeupInterval
}

func (s *Scheduler) wakeup() {
	s.timer.Reset(0)
}
