package agscheduler

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/kwkwc/agscheduler/services/proto"
)

type Scheduler struct {
	store     Store
	timer     *time.Timer
	quitChan  chan struct{}
	isRunning bool

	clusterNode *ClusterNode
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

func (s *Scheduler) SetClusterNode(cn *ClusterNode) error {
	s.clusterNode = cn
	if err := s.clusterNode.init(); err != nil {
		return err
	}

	return nil
}

func CalcNextRunTime(j Job) (time.Time, error) {
	timezone, err := time.LoadLocation(j.Timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("job `%s` Timezone `%s` error: %s", j.FullName(), j.Timezone, err)
	}

	if j.Status == STATUS_PAUSED {
		nextRunTimeMax, _ := time.ParseInLocation(time.DateTime, "9999-09-09 09:09:09", timezone)
		return time.Unix(nextRunTimeMax.Unix(), 0).UTC(), nil
	}

	var nextRunTime time.Time
	switch strings.ToLower(j.Type) {
	case TYPE_DATETIME:
		nextRunTime, err = time.ParseInLocation(time.DateTime, j.StartAt, timezone)
		if err != nil {
			return time.Time{}, fmt.Errorf("job `%s` StartAt `%s` error: %s", j.FullName(), j.Timezone, err)
		}
	case TYPE_INTERVAL:
		i, err := time.ParseDuration(j.Interval)
		if err != nil {
			return time.Time{}, fmt.Errorf("job `%s` Interval `%s` error: %s", j.FullName(), j.Interval, err)
		}
		nextRunTime = time.Now().In(timezone).Add(i)
	case TYPE_CRON:
		nextRunTime = cronexpr.MustParse(j.CronExpr).Next(time.Now().In(timezone))
	default:
		return time.Time{}, fmt.Errorf("job `%s` Type `%s` unknown", j.FullName(), j.Type)
	}

	return time.Unix(nextRunTime.Unix(), 0).UTC(), nil
}

func (s *Scheduler) AddJob(j Job) (Job, error) {
	for {
		j.setId()
		if _, err := s.GetJob(j.Id); err != nil {
			break
		}
	}

	slog.Info(fmt.Sprintf("Scheduler add job `%s`.\n", j.FullName()))

	j.Status = STATUS_RUNNING

	if j.Timezone == "" {
		j.Timezone = "UTC"
	}

	if j.FuncName == "" {
		j.FuncName = getFuncName(j.Func)
	}
	if _, ok := funcMap[j.FuncName]; !ok {
		return Job{}, FuncUnregisteredError(j.FuncName)
	}

	nextRunTime, err := CalcNextRunTime(j)
	if err != nil {
		return Job{}, err
	}
	j.NextRunTime = nextRunTime

	if err := s.store.AddJob(j); err != nil {
		return Job{}, err
	}

	if !s.isRunning {
		s.Start()
	}

	return j, nil
}

func (s *Scheduler) GetJob(id string) (Job, error) {
	return s.store.GetJob(id)
}

func (s *Scheduler) GetAllJobs() ([]Job, error) {
	return s.store.GetAllJobs()
}

func (s *Scheduler) UpdateJob(j Job) (Job, error) {
	if _, err := s.GetJob(j.Id); err != nil {
		return Job{}, err
	}

	lastNextWakeupInterval := s.getNextWakeupInterval()

	if _, ok := funcMap[j.FuncName]; !ok {
		return Job{}, FuncUnregisteredError(j.FuncName)
	}

	nextRunTime, err := CalcNextRunTime(j)
	if err != nil {
		return Job{}, err
	}
	j.NextRunTime = nextRunTime

	err = s.store.UpdateJob(j)

	nextWakeupInterval := s.getNextWakeupInterval()
	if nextWakeupInterval < lastNextWakeupInterval {
		s.wakeup()
	}

	return j, err
}

func (s *Scheduler) DeleteJob(id string) error {
	slog.Info(fmt.Sprintf("Scheduler delete jobId `%s`.\n", id))

	if _, err := s.GetJob(id); err != nil {
		return err
	}

	return s.store.DeleteJob(id)
}

func (s *Scheduler) DeleteAllJobs() error {
	slog.Info("Scheduler delete all jobs.\n")

	return s.store.DeleteAllJobs()
}

func (s *Scheduler) PauseJob(id string) (Job, error) {
	slog.Info(fmt.Sprintf("Scheduler pause jobId `%s`.\n", id))

	j, err := s.GetJob(id)
	if err != nil {
		return Job{}, err
	}

	j.Status = STATUS_PAUSED

	j, err = s.UpdateJob(j)
	if err != nil {
		return Job{}, err
	}

	return j, nil
}

func (s *Scheduler) ResumeJob(id string) (Job, error) {
	slog.Info(fmt.Sprintf("Scheduler resume jobId `%s`.\n", id))

	j, err := s.GetJob(id)
	if err != nil {
		return Job{}, err
	}

	j.Status = STATUS_RUNNING

	j, err = s.UpdateJob(j)
	if err != nil {
		return Job{}, err
	}

	return j, nil
}

func (s *Scheduler) _runJob(j Job) {
	f := reflect.ValueOf(funcMap[j.FuncName])
	if f.IsNil() {
		slog.Warn(fmt.Sprintf("Job `%s` Func `%s` unregistered\n", j.FullName(), j.FuncName))
	} else {
		slog.Info(fmt.Sprintf("Job `%s` is running, next run time: `%s`\n", j.FullName(), j.NextRunTimeWithTimezone().String()))
		go func() {
			defer func() {
				if err := recover(); err != nil {
					slog.Error(fmt.Sprintf("Job `%s` panic: %s\n", j.FullName(), err))
					slog.Debug(fmt.Sprintf("%s\n", string(debug.Stack())))
				}
			}()

			f.Call([]reflect.Value{reflect.ValueOf(j)})
		}()
	}
}

func (s *Scheduler) _runJobRemote(node *ClusterNode, j Job) {
	go func() {
		conn, _ := grpc.Dial(node.SchedulerEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		defer conn.Close()

		client := pb.NewSchedulerClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, err := client.RunJob(ctx, JobToPbJobPtr(j))
		if err != nil {
			slog.Error(fmt.Sprintf("Scheduler run job remote error %s\n", err))
			s.clusterNode.queueMap[node.SchedulerQueue][node.Id]["health"] = false
		}
	}()
}

func (s *Scheduler) _flushJob(j Job, now time.Time) error {
	j.LastRunTime = time.Unix(now.Unix(), 0).UTC()

	if j.Type == TYPE_DATETIME {
		if j.NextRunTime.Before(now) {
			if err := s.DeleteJob(j.Id); err != nil {
				return fmt.Errorf("delete job `%s` error: %s", j.FullName(), err)
			}
		}
	} else {
		if _, err := s.UpdateJob(j); err != nil {
			return fmt.Errorf("update job `%s` error: %s", j.FullName(), err)
		}
	}

	return nil
}

func (s *Scheduler) RunJob(j Job) error {
	slog.Info(fmt.Sprintf("Scheduler run job `%s`.\n", j.FullName()))

	s._runJob(j)

	return nil
}

func (s *Scheduler) scheduleJob(j Job) error {
	if s.clusterNode == nil {
		s._runJob(j)
	} else {
		node, err := s.clusterNode.choiceNode()
		if err != nil {
			s._runJob(j)
		} else {
			s._runJobRemote(node, j)
		}
	}

	return nil
}

func (s *Scheduler) run() {
	for {
		select {
		case <-s.quitChan:
			slog.Info("Scheduler quit.\n")
			return
		case <-s.timer.C:
			now := time.Now().UTC()

			js, err := s.GetAllJobs()
			if err != nil {
				slog.Error(fmt.Sprintf("Scheduler get all jobs error: %s\n", err))
				continue
			}
			if len(js) == 0 {
				s.Stop()
				continue
			}
			sort.Sort(JobSlice(js))

			for _, j := range js {
				if j.NextRunTime.Before(now) {
					nextRunTime, err := CalcNextRunTime(j)
					if err != nil {
						slog.Error(fmt.Sprintf("Scheduler calc next run time error: %s\n", err))
						continue
					}
					j.NextRunTime = nextRunTime

					err = s.scheduleJob(j)
					if err != nil {
						slog.Error(fmt.Sprintf("Scheduler schedule job error %s\n", err))
					}

					err = s._flushJob(j, now)
					if err != nil {
						slog.Error(fmt.Sprintf("Scheduler %s\n", err))
						continue
					}
				} else {
					break
				}
			}

			nextWakeupInterval := s.getNextWakeupInterval()
			slog.Debug(fmt.Sprintf("Scheduler next wakeup interval %s\n", nextWakeupInterval))

			s.timer.Reset(nextWakeupInterval)
		}
	}
}

func (s *Scheduler) Start() {
	if s.isRunning {
		slog.Info("Scheduler is running.\n")
		return
	}

	s.timer = time.NewTimer(0)
	s.quitChan = make(chan struct{}, 3)
	s.isRunning = true

	go s.run()

	slog.Info("Scheduler start.\n")
}

func (s *Scheduler) Stop() {
	if !s.isRunning {
		slog.Info("Scheduler has stopped.\n")
		return
	}

	s.quitChan <- struct{}{}
	s.isRunning = false

	slog.Info("Scheduler stop.\n")
}

func (s *Scheduler) getNextWakeupInterval() time.Duration {
	nextRunTimeMin, err := s.store.GetNextRunTime()
	if err != nil {
		slog.Error(fmt.Sprintf("Scheduler get next wakeup interval error: %s\n", err))
		nextRunTimeMin = time.Now().UTC().Add(1 * time.Second)
	}

	now := time.Now().UTC()
	nextWakeupInterval := nextRunTimeMin.Sub(now)
	if nextWakeupInterval < 0 {
		nextWakeupInterval = time.Second
	}

	return nextWakeupInterval
}

func (s *Scheduler) wakeup() {
	s.timer.Reset(0)
}
