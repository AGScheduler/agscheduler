package agscheduler

import (
	"context"
	"fmt"
	"log/slog"
	"net/rpc"
	"reflect"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorhill/cronexpr"
)

var GetStore = (*Scheduler).getStore
var GetClusterNode = (*Scheduler).getClusterNode
var GetBroker = (*Scheduler).getBroker

var mutexS sync.RWMutex

// In standalone mode, the scheduler only needs to run jobs on a regular basis.
// In cluster mode, the scheduler also needs to be responsible for allocating jobs to cluster nodes.
type Scheduler struct {
	// Job store
	store Store
	// When the time is up, the scheduler will wake up.
	timer *time.Timer
	// Input is received when `stop` is called or no job in store.
	quitChan chan struct{}
	// It should not be set manually.
	isRunning bool

	// Used in cluster mode, bind to each other and the cluster node.
	clusterNode *ClusterNode

	// Used in broker mode, bind to each other and the broker.
	broker *Broker
}

func (s *Scheduler) IsRunning() bool {
	mutexS.RLock()
	defer mutexS.RUnlock()

	return s.isRunning
}

// Bind the store
func (s *Scheduler) SetStore(sto Store) error {
	s.store = sto
	if err := s.store.Init(); err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) getStore() Store {
	return s.store
}

// Bind the cluster node
func (s *Scheduler) SetClusterNode(ctx context.Context, cn *ClusterNode) error {
	s.clusterNode = cn
	cn.Scheduler = s
	if err := s.clusterNode.init(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) IsClusterMode() bool {
	return s.clusterNode != nil
}

func (s *Scheduler) getClusterNode() *ClusterNode {
	return s.clusterNode
}

// Bind the broker
func (s *Scheduler) SetBroker(brk *Broker) error {
	s.broker = brk
	brk.scheduler = s
	if err := s.broker.init(); err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) getBroker() *Broker {
	return s.broker
}

func (s *Scheduler) IsBrokerMode() bool {
	return s.broker != nil
}

// Calculate the next run time, different job type will be calculated in different ways,
// when the job is paused, will return `9999-09-09 09:09:09`.
func CalcNextRunTime(j Job) (time.Time, error) {
	timezone, err := time.LoadLocation(j.Timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("job `%s` Timezone `%s` error: %s", j.FullName(), j.Timezone, err)
	}

	if j.Status == STATUS_PAUSED {
		nextRunTimeMax, _ := GetNextRunTimeMax()
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
		expr, err := cronexpr.Parse(j.CronExpr)
		if err != nil {
			return time.Time{}, fmt.Errorf("job `%s` CronExpr `%s` error: %s", j.FullName(), j.CronExpr, err)
		}
		nextRunTime = expr.Next(time.Now().In(timezone))
	default:
		return time.Time{}, fmt.Errorf("job `%s` Type `%s` unknown", j.FullName(), j.Type)
	}

	return time.Unix(nextRunTime.Unix(), 0).UTC(), nil
}

func (s *Scheduler) AddJob(j Job) (Job, error) {
	if err := j.init(); err != nil {
		return Job{}, err
	}

	slog.Info(fmt.Sprintf("Scheduler add job `%s`.", j.FullName()))

	if err := s.store.AddJob(j); err != nil {
		return Job{}, err
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
	oJ, err := s.GetJob(j.Id)
	if err != nil {
		return Job{}, err
	}

	if j.Status == "" ||
		(j.Status != STATUS_RUNNING && j.Status != STATUS_PAUSED) {
		j.Status = oJ.Status
	}

	if err := j.check(); err != nil {
		return Job{}, err
	}

	nextRunTime, err := CalcNextRunTime(j)
	if err != nil {
		return Job{}, err
	}
	j.NextRunTime = nextRunTime

	lastNextWakeupInterval := s.getNextWakeupInterval()

	if err := s.store.UpdateJob(j); err != nil {
		return Job{}, err
	}

	nextWakeupInterval := s.getNextWakeupInterval()
	if nextWakeupInterval < lastNextWakeupInterval {
		s.wakeup()
	}

	return j, nil
}

func (s *Scheduler) DeleteJob(id string) error {
	slog.Info(fmt.Sprintf("Scheduler delete jobId `%s`.", id))

	if _, err := s.GetJob(id); err != nil {
		return err
	}

	return s.store.DeleteJob(id)
}

func (s *Scheduler) DeleteAllJobs() error {
	slog.Info("Scheduler delete all jobs.")

	return s.store.DeleteAllJobs()
}

func (s *Scheduler) PauseJob(id string) (Job, error) {
	slog.Info(fmt.Sprintf("Scheduler pause jobId `%s`.", id))

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
	slog.Info(fmt.Sprintf("Scheduler resume jobId `%s`.", id))

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

// Used in broker mode.
// Push job to queue to run the `RunJob`.
func (s *Scheduler) pushJob(queue string, j Job) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("Job `%s` push to queue:`%s` error: %s", j.FullName(), queue, err))
			slog.Debug(string(debug.Stack()))
		}
	}()

	if err := s.broker.Queues[queue].PushJob(j); err != nil {
		panic(err)
	}
}

// Used in standalone mode.
func (s *Scheduler) _runJob(j Job) {
	f := reflect.ValueOf(FuncMap[j.FuncName].Func)
	if f.IsNil() {
		slog.Warn(fmt.Sprintf("Job `%s` Func `%s` unregistered", j.FullName(), j.FuncName))
	} else {
		slog.Info(fmt.Sprintf("Job `%s` is running, next run time: `%s`", j.FullName(), j.NextRunTimeWithTimezone().String()))

		timeout, err := time.ParseDuration(j.Timeout)
		if err != nil {
			e := &JobTimeoutError{FullName: j.FullName(), Timeout: j.Timeout, Err: err}
			slog.Error(e.Error())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		ch := make(chan error, 1)
		go func() {
			defer close(ch)
			defer func() {
				if err := recover(); err != nil {
					slog.Error(fmt.Sprintf("Job `%s` run error: %s", j.FullName(), err))
					slog.Debug(string(debug.Stack()))
				}
			}()

			f.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(j)})
		}()

		select {
		case <-ch:
			return
		case <-ctx.Done():
			slog.Warn(fmt.Sprintf("Job `%s` run timeout", j.FullName()))
		}
	}
}

// Used in cluster mode.
// Call the RPC API of the other node to run the `RunJob`.
func (s *Scheduler) _runJobRemote(node *ClusterNode, j Job) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("Job `%s` _runJobRemote error: %s", j.FullName(), err))
			slog.Debug(string(debug.Stack()))
		}
	}()

	rClient, err := rpc.DialHTTP("tcp", node.Endpoint)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to cluster node: `%s`, error: %s", node.Endpoint, err))
		return
	}
	defer rClient.Close()

	var r any
	ch := make(chan error, 1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				slog.Error(fmt.Sprintf("Job `%s` CRPCService.RunJob error: %s", j.FullName(), err))
				slog.Debug(string(debug.Stack()))
			}
		}()

		ch <- rClient.Call("CRPCService.RunJob", j, &r)
	}()
	select {
	case err := <-ch:
		if err != nil {
			slog.Error(fmt.Sprintf("Scheduler run job `%s` remote error %s", j.FullName(), err))
		}
	case <-time.After(3 * time.Second):
		slog.Error(fmt.Sprintf("Scheduler run job `%s` remote timeout %s", j.FullName(), err))
	}
}

func (s *Scheduler) _flushJob(j Job, now time.Time) error {
	if j.Type == TYPE_DATETIME {
		if j.NextRunTime.Before(now) {
			if err := s.DeleteJob(j.Id); err != nil {
				return fmt.Errorf("delete job `%s` error: %s", j.FullName(), err)
			}
		}
	} else {
		j, err := s.GetJob(j.Id)
		if err != nil {
			return fmt.Errorf("get job `%s` error: %s", j.FullName(), err)
		}
		j.LastRunTime = time.Unix(now.Unix(), 0).UTC()
		if _, err := s.UpdateJob(j); err != nil {
			return fmt.Errorf("update job `%s` error: %s", j.FullName(), err)
		}
	}

	return nil
}

func (s *Scheduler) _scheduleJob(j Job) error {
	if s.IsClusterMode() {
		// In cluster mode, all nodes are equal and may pick myself.
		node, err := s.clusterNode.choiceNode(j.Queues)
		if err != nil {
			return fmt.Errorf("cluster node with queue `%s` does not exist", j.Queues)
		}
		go s._runJobRemote(node, j)
	} else {
		// In standalone mode.
		if s.IsBrokerMode() {
			// In broker mode.
			queue, err := s.broker.choiceQueue(j.Queues)
			if err != nil {
				return fmt.Errorf("broker's queues with queue `%s` does not exist", j.Queues)
			}
			go s.pushJob(queue, j)
		} else {
			go s._runJob(j)
		}
	}

	return nil
}

func (s *Scheduler) RunJob(j Job) error {
	slog.Info(fmt.Sprintf("Scheduler run job `%s`.", j.FullName()))

	go s._runJob(j)

	return nil
}

// Used in cluster mode.
// Select a worker node
func (s *Scheduler) ScheduleJob(j Job) error {
	slog.Info(fmt.Sprintf("Scheduler schedule job `%s`.", j.FullName()))

	err := s._scheduleJob(j)
	if err != nil {
		return fmt.Errorf("scheduler schedule job `%s` error: %s", j.FullName(), err)
	}

	return nil
}

func (s *Scheduler) run() {
	for {
		select {
		case <-s.quitChan:
			slog.Info("Scheduler quit.")
			return
		case <-s.timer.C:
			if s.IsClusterMode() && !s.clusterNode.IsMainNode() {
				s.timer.Reset(time.Second)
				continue
			}

			now := time.Now().UTC()

			js, err := s.GetAllJobs()
			if err != nil {
				slog.Error(fmt.Sprintf("Scheduler get all jobs error: %s", err))
				s.timer.Reset(time.Second)
				continue
			}

			// If there are ineligible job, subsequent job do not need to be checked.
			sort.Sort(JobSlice(js))
			for _, j := range js {
				if j.NextRunTime.Before(now) {
					nextRunTime, err := CalcNextRunTime(j)
					if err != nil {
						slog.Error(fmt.Sprintf("Scheduler calc next run time error: %s", err))
						continue
					}
					j.NextRunTime = nextRunTime

					err = s._scheduleJob(j)
					if err != nil {
						slog.Error(fmt.Sprintf("Scheduler schedule job `%s` error: %s", j.FullName(), err))
					}

					err = s._flushJob(j, now)
					if err != nil {
						slog.Error(fmt.Sprintf("Scheduler %s", err))
						continue
					}
				} else {
					break
				}
			}

			nextWakeupInterval := s.getNextWakeupInterval()
			slog.Debug(fmt.Sprintf("Scheduler next wakeup interval %s", nextWakeupInterval))

			s.timer.Reset(nextWakeupInterval)
		}
	}
}

// In addition to being called manually,
// it is also called after `AddJob`.
func (s *Scheduler) Start() {
	mutexS.Lock()
	defer mutexS.Unlock()

	if s.isRunning {
		slog.Info("Scheduler is running.")
		return
	}

	s.timer = time.NewTimer(0)
	s.quitChan = make(chan struct{}, 3)
	s.isRunning = true

	go s.run()

	slog.Info("Scheduler start.")
}

// In addition to being called manually,
// there is no job in store that will also be called.
func (s *Scheduler) Stop() {
	mutexS.Lock()
	defer mutexS.Unlock()

	if !s.isRunning {
		slog.Info("Scheduler has stopped.")
		return
	}

	s.quitChan <- struct{}{}
	s.isRunning = false

	slog.Info("Scheduler stop.")
}

// Dynamically calculate the next wakeup interval, avoid frequent wakeup of the scheduler
func (s *Scheduler) getNextWakeupInterval() time.Duration {
	nextRunTimeMin, err := s.store.GetNextRunTime()
	if err != nil {
		slog.Error(fmt.Sprintf("Scheduler get next wakeup interval error: %s", err))
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
	if s.timer != nil {
		s.timer.Reset(0)
	}
}

func (s *Scheduler) Info() map[string]any {
	info := map[string]any{
		"cluster_main_node": map[string]any{},
		"is_cluster_mode":   s.IsClusterMode(),
		"is_running":        s.IsRunning(),
		"version":           Version,
	}

	if s.IsClusterMode() {
		info["cluster_main_node"] = map[string]any{
			"endpoint_main": s.clusterNode.GetEndpointMain(),
			"endpoint":      s.clusterNode.Endpoint,
			"endpoint_grpc": s.clusterNode.EndpointGRPC,
			"endpoint_http": s.clusterNode.EndpointHTTP,
			"mode":          s.clusterNode.Mode,
		}
	}

	return info
}
