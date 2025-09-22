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
var GetRecorder = (*Scheduler).getRecorder
var GetListener = (*Scheduler).getListener

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

	// When broker exist, job scheduling is done in queue.
	broker *Broker
	// When recorder exist, record the results of job runs.
	recorder *Recorder
	listener *Listener

	// Track running job instances for max_instances control
	runningJobs   map[string]int
	jobInstancesM sync.RWMutex

	statusM sync.RWMutex
	storeM  sync.RWMutex
}

func (s *Scheduler) IsRunning() bool {
	s.statusM.RLock()
	defer s.statusM.RUnlock()

	return s.isRunning
}

// Bind the store
func (s *Scheduler) SetStore(sto Store) error {
	slog.Info("Scheduler set Store.")

	s.store = sto
	if err := s.store.Init(); err != nil {
		return err
	}

	s.init()
	return nil
}

func (s *Scheduler) getStore() Store {
	return s.store
}

func (s *Scheduler) init() {
	s.runningJobs = make(map[string]int)
}

func (s *Scheduler) canRunJob(jobName string, maxInstances int) bool {
	s.jobInstancesM.RLock()
	defer s.jobInstancesM.RUnlock()

	return s.runningJobs[jobName] < maxInstances
}

func (s *Scheduler) incrementJobInstance(jobName string) {
	s.jobInstancesM.Lock()
	defer s.jobInstancesM.Unlock()

	s.runningJobs[jobName]++
}

func (s *Scheduler) decrementJobInstance(jobName string) {
	s.jobInstancesM.Lock()
	defer s.jobInstancesM.Unlock()

	if s.runningJobs[jobName] > 0 {
		s.runningJobs[jobName]--
	}
}

func (s *Scheduler) getJobInstanceCount(jobName string) int {
	s.jobInstancesM.RLock()
	defer s.jobInstancesM.RUnlock()

	return s.runningJobs[jobName]
}

// Bind the cluster node
func (s *Scheduler) SetClusterNode(ctx context.Context, cn *ClusterNode) error {
	slog.Info("Scheduler set ClusterNode.")

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
func (s *Scheduler) SetBroker(ctx context.Context, brk *Broker) error {
	slog.Info("Scheduler set Broker.")

	s.broker = brk
	brk.scheduler = s
	if err := s.broker.init(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) getBroker() *Broker {
	return s.broker
}

func (s *Scheduler) HasBroker() bool {
	return s.broker != nil
}

// Bind the recorder
func (s *Scheduler) SetRecorder(rec *Recorder) error {
	slog.Info("Scheduler set Recorder.")

	s.recorder = rec
	if err := s.recorder.init(); err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) getRecorder() *Recorder {
	return s.recorder
}

func (s *Scheduler) HasRecorder() bool {
	return s.recorder != nil
}

// Bind the listener
func (s *Scheduler) SetListener(lis *Listener) error {
	slog.Info("Scheduler set Listener.")

	s.listener = lis
	if err := s.listener.init(); err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) getListener() *Listener {
	return s.listener
}

func (s *Scheduler) HasListener() bool {
	return s.listener != nil
}

// Calculate the next run time, different job type will be calculated in different ways,
// when the job is paused, will return `9999-09-09 09:09:09`.
func CalcNextRunTime(j Job) (time.Time, error) {
	timezone, err := time.LoadLocation(j.Timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("job `%s` Timezone `%s` error: %s", j.FullName(), j.Timezone, err)
	}

	if j.Status == JOB_STATUS_PAUSED {
		nextRunTimeMax, _ := GetNextRunTimeMax()
		return time.Unix(nextRunTimeMax.Unix(), 0).UTC(), nil
	}

	var nextRunTime time.Time
	switch strings.ToLower(j.Type) {
	case JOB_TYPE_DATETIME:
		nextRunTime, err = time.ParseInLocation(time.DateTime, j.StartAt, timezone)
		if err != nil {
			return time.Time{}, fmt.Errorf("job `%s` StartAt `%s` error: %s", j.FullName(), j.Timezone, err)
		}
	case JOB_TYPE_INTERVAL:
		i, err := time.ParseDuration(j.Interval)
		if err != nil {
			return time.Time{}, fmt.Errorf("job `%s` Interval `%s` error: %s", j.FullName(), j.Interval, err)
		}
		nextRunTime = time.Now().In(timezone).Add(i)
	case JOB_TYPE_CRON:
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
	s.storeM.Lock()
	defer s.storeM.Unlock()

	if err := j.init(); err != nil {
		return Job{}, err
	}

	slog.Info(fmt.Sprintf("Scheduler add job `%s`.", j.FullName()))

	lastNextWakeupInterval := s.getNextWakeupInterval()

	if err := s.store.AddJob(j); err != nil {
		return Job{}, err
	}

	nextWakeupInterval := s.getNextWakeupInterval()
	if nextWakeupInterval < lastNextWakeupInterval {
		s.wakeup()
	}

	s.dispatchEvent(EventPkg{EVENT_JOB_ADDED, j.Id, nil})
	return j, nil
}

func (s *Scheduler) GetJob(id string) (Job, error) {
	s.storeM.RLock()
	defer s.storeM.RUnlock()

	return s.store.GetJob(id)
}

func (s *Scheduler) GetAllJobs() ([]Job, error) {
	s.storeM.RLock()
	defer s.storeM.RUnlock()

	return s.store.GetAllJobs()
}

func (s *Scheduler) _updateJob(j Job) (Job, error) {
	oJ, err := s.store.GetJob(j.Id)
	if err != nil {
		return Job{}, err
	}

	if j.Status == "" ||
		(j.Status != JOB_STATUS_RUNNING && j.Status != JOB_STATUS_PAUSED) {
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

	s.dispatchEvent(EventPkg{EVENT_JOB_UPDATED, j.Id, nil})
	return j, nil
}

func (s *Scheduler) UpdateJob(j Job) (Job, error) {
	s.storeM.Lock()
	defer s.storeM.Unlock()

	j, err := s._updateJob(j)
	if err != nil {
		return Job{}, err
	}

	return j, nil
}

func (s *Scheduler) _deleteJob(id string) error {
	slog.Info(fmt.Sprintf("Scheduler delete jobId `%s`.", id))

	if _, err := s.store.GetJob(id); err != nil {
		return err
	}

	if err := s.store.DeleteJob(id); err != nil {
		return err
	}

	s.dispatchEvent(EventPkg{EVENT_JOB_DELETED, id, nil})
	return nil
}

func (s *Scheduler) DeleteJob(id string) error {
	s.storeM.Lock()
	defer s.storeM.Unlock()

	return s._deleteJob(id)
}

func (s *Scheduler) DeleteAllJobs() error {
	s.storeM.Lock()
	defer s.storeM.Unlock()

	slog.Info("Scheduler delete all jobs.")

	if err := s.store.DeleteAllJobs(); err != nil {
		return err
	}

	s.dispatchEvent(EventPkg{EVENT_ALL_JOBS_DELETED, "", nil})
	return nil
}

func (s *Scheduler) PauseJob(id string) (Job, error) {
	s.storeM.Lock()
	defer s.storeM.Unlock()

	slog.Info(fmt.Sprintf("Scheduler pause jobId `%s`.", id))

	j, err := s.store.GetJob(id)
	if err != nil {
		return Job{}, err
	}

	j.Status = JOB_STATUS_PAUSED

	j, err = s._updateJob(j)
	if err != nil {
		return Job{}, err
	}

	s.dispatchEvent(EventPkg{EVENT_JOB_PAUSED, j.Id, nil})
	return j, nil
}

func (s *Scheduler) ResumeJob(id string) (Job, error) {
	s.storeM.Lock()
	defer s.storeM.Unlock()

	slog.Info(fmt.Sprintf("Scheduler resume jobId `%s`.", id))

	j, err := s.store.GetJob(id)
	if err != nil {
		return Job{}, err
	}

	j.Status = JOB_STATUS_RUNNING

	j, err = s._updateJob(j)
	if err != nil {
		return Job{}, err
	}

	s.dispatchEvent(EventPkg{EVENT_JOB_RESUMED, j.Id, nil})
	return j, nil
}

// When broker exist, push job to queue to run the `RunJob`.
func (s *Scheduler) pushJob(queue string, j Job) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("Job `%s` push to queue:`%s` error: %s", j.FullName(), queue, err))
			slog.Debug(string(debug.Stack()))
		}
	}()

	bJ, err := JobMarshal(j)
	if err != nil {
		panic(err)
	}
	if err := s.broker.pushJob(queue, bJ); err != nil {
		panic(err)
	}
}

// Used in standalone mode.
func (s *Scheduler) _runJob(j Job) {
	if !s.canRunJob(j.Name, j.MaxInstances) {
		slog.Warn(fmt.Sprintf("Job `%s` skipped due to max_instances limit (%d/%d)", j.FullName(), s.getJobInstanceCount(j.Name), j.MaxInstances))
		s.dispatchEvent(EventPkg{EVENT_JOB_MAX_INSTANCES, j.Id, nil})
		return
	}

	s.incrementJobInstance(j.Name)
	defer s.decrementJobInstance(j.Name)

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

		var rId uint64
		var status string
		var result string
		if s.HasRecorder() {
			rId, err = s.recorder.RecordMetadata(j)
			if err != nil {
				slog.Error(fmt.Sprintf("Job `%s` record metadata error: `%s`", j.FullName(), err))
				return
			}
		}

		ch := make(chan error, 1)
		go func() {
			defer close(ch)
			defer func() {
				if err := recover(); err != nil {
					slog.Error(fmt.Sprintf("Job `%s` run error: %s", j.FullName(), err))
					s.dispatchEvent(EventPkg{EVENT_JOB_ERROR, j.Id, err})
					slog.Debug(string(debug.Stack()))
					status = RECORD_STATUS_ERROR
					result = fmt.Sprintf("%s", err)
				}
			}()

			rValues := f.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(j)})
			result = rValues[0].Interface().(string)
		}()

		select {
		case <-ch:
			s.dispatchEvent(EventPkg{EVENT_JOB_EXECUTED, j.Id, nil})
			if status == "" {
				status = RECORD_STATUS_COMPLETED
			}
		case <-ctx.Done():
			slog.Warn(fmt.Sprintf("Job `%s` run timeout", j.FullName()))
			s.dispatchEvent(EventPkg{EVENT_JOB_TIMEOUT, j.Id, nil})
			status = RECORD_STATUS_TIMEOUT
		}

		if s.HasRecorder() {
			err := s.recorder.RecordResult(rId, status, result)
			if err != nil {
				slog.Error(fmt.Sprintf("Job `%s` record result error: `%s`", j.FullName(), err))
				return
			}
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
	if j.Type == JOB_TYPE_DATETIME {
		if j.NextRunTime.Before(now) {
			if err := s._deleteJob(j.Id); err != nil {
				return fmt.Errorf("delete job `%s` error: %s", j.FullName(), err)
			}
		}
	} else {
		j, err := s.store.GetJob(j.Id)
		if err != nil {
			return fmt.Errorf("get job `%s` error: %s", j.FullName(), err)
		}
		j.LastRunTime = time.Unix(now.Unix(), 0).UTC()
		if _, err := s._updateJob(j); err != nil {
			return fmt.Errorf("update job `%s` error: %s", j.FullName(), err)
		}
	}

	return nil
}

// All nodes are equal and may pick myself.
func (s *Scheduler) _scheduleJob(j Job) error {
	if s.HasBroker() {
		// When broker exist.
		queue, err := s.broker.choiceQueue(j.Queues)
		if err != nil {
			return fmt.Errorf("broker's queues with queue `%s` does not exist", j.Queues)
		}
		go s.pushJob(queue, j)
	} else {
		if s.IsClusterMode() {
			// In cluster mode.
			node, err := s.clusterNode.choiceNode(j.Queues)
			if err != nil {
				return fmt.Errorf("cluster node with queue `%s` does not exist", j.Queues)
			}
			go s._runJobRemote(node, j)
		} else {
			// In standalone mode.
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

// Select a worker node or queue.
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
			s.timer.Stop()

			if s.IsClusterMode() && !s.clusterNode.IsMainNode() {
				s.timer.Reset(time.Second)
				continue
			}

			now := time.Now().UTC()

			s.storeM.Lock()
			js, err := s.store.GetAllJobs()
			if err != nil {
				slog.Error(fmt.Sprintf("Scheduler get all jobs error: %s", err))
				s.timer.Reset(time.Second)
				s.storeM.Unlock()
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
			s.storeM.Unlock()
		}
	}
}

// In addition to being called manually,
// it is also called after `AddJob`.
func (s *Scheduler) Start() {
	s.statusM.Lock()
	defer s.statusM.Unlock()

	if s.isRunning {
		slog.Info("Scheduler is running.")
		return
	}

	s.timer = time.NewTimer(0)
	s.quitChan = make(chan struct{})
	s.isRunning = true

	go s.run()

	slog.Info("Scheduler start.")
	s.dispatchEvent(EventPkg{EVENT_SCHEDULER_STARTED, "", nil})
}

// In addition to being called manually,
// there is no job in store that will also be called.
func (s *Scheduler) Stop() {
	s.statusM.Lock()
	defer s.statusM.Unlock()

	if !s.isRunning {
		slog.Info("Scheduler has stopped.")
		return
	}

	s.quitChan <- struct{}{}
	s.isRunning = false

	slog.Info("Scheduler stop.")
	s.dispatchEvent(EventPkg{EVENT_SCHEDULER_STOPPED, "", nil})
}

// Dynamically calculate the next wakeup interval, avoid frequent wakeup of the scheduler.
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

// Send an event to the listener.
func (s *Scheduler) dispatchEvent(eP EventPkg) {
	if !s.HasListener() {
		return
	}
	s.listener.handleEvent(eP)
}

func (s *Scheduler) Info() map[string]any {
	info := map[string]any{
		"scheduler": map[string]any{
			"is_running": s.IsRunning(),
			"store":      s.store.Name(),
		},
		"broker": map[string]any{
			"has_broker": s.HasBroker(),
			"queues":     "",
		},
		"recorder": map[string]any{
			"has_recorder": s.HasRecorder(),
			"backend":      "",
		},
		"cluster": map[string]any{
			"is_cluster_mode": s.IsClusterMode(),
			"main_node":       map[string]any{},
		},
		"version": Version,
	}

	if s.HasBroker() {
		queues := []string{}
		for k := range s.broker.Queues {
			queues = append(queues, k)
		}
		info["broker"].(map[string]any)["queues"] = strings.Join(queues, ",")
	}

	if s.HasRecorder() {
		info["recorder"].(map[string]any)["backend"] = s.recorder.Backend.Name()
	}

	if s.IsClusterMode() {
		info["cluster"].(map[string]any)["main_node"] = map[string]any{
			"endpoint_main": s.clusterNode.GetEndpointMain(),
			"endpoint":      s.clusterNode.Endpoint,
			"endpoint_grpc": s.clusterNode.EndpointGRPC,
			"endpoint_http": s.clusterNode.EndpointHTTP,
			"mode":          s.clusterNode.Mode,
		}
	}

	return info
}
