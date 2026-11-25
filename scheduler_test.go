package agscheduler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gorhill/cronexpr"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/internal/adaptertest"
)

func dryRunScheduler(ctx context.Context, j agscheduler.Job) (result string) { return }

func runSchedulerPanic(ctx context.Context, j agscheduler.Job) (result string) { panic(nil) }

func runSchedulerWithMetadataFailure(ctx context.Context, j agscheduler.Job) (result string) {
	j.Args["ran"].(chan struct{}) <- struct{}{}
	return "ran"
}

func dryCallbackScheduler(ep agscheduler.EventPkg) {}

type failingMetadataBackend struct {
	agscheduler.Backend
}

func (b *failingMetadataBackend) RecordMetadata(agscheduler.Record) error {
	return errors.New("metadata backend unavailable")
}

type failingInitStore struct {
	agscheduler.Store
}

func (s *failingInitStore) Init() error {
	return errors.New("store init failed")
}

type failingInitBackend struct {
	agscheduler.Backend
}

func (b *failingInitBackend) Init() error {
	return errors.New("backend init failed")
}

var gracefulSchedulerStarted chan struct{}
var gracefulSchedulerRelease chan struct{}
var cancellationSchedulerStarted chan struct{}
var cancellationSchedulerObserved chan struct{}
var preparedSchedulerJob chan agscheduler.Job

func waitForSchedulerRelease(ctx context.Context, j agscheduler.Job) (result string) {
	gracefulSchedulerStarted <- struct{}{}
	<-gracefulSchedulerRelease
	return "released"
}

func waitForSchedulerCancellation(ctx context.Context, j agscheduler.Job) (result string) {
	cancellationSchedulerStarted <- struct{}{}
	<-ctx.Done()
	close(cancellationSchedulerObserved)
	return ""
}

func capturePreparedSchedulerJob(ctx context.Context, j agscheduler.Job) (result string) {
	preparedSchedulerJob <- j
	return ""
}

func getSchedulerWithStore(t *testing.T) *agscheduler.Scheduler {
	scheduler := &agscheduler.Scheduler{}
	store := adaptertest.NewStore()
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	return scheduler
}

func getJob() agscheduler.Job {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: dryRunScheduler},
		agscheduler.FuncPkg{Func: runSchedulerPanic},
	)

	return agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "50ms",
		Func:     dryRunScheduler,
	}
}

func TestSchedulerRunJobContinuesWhenRecordMetadataFails(t *testing.T) {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: runSchedulerWithMetadataFailure},
	)

	scheduler := getSchedulerWithStore(t)
	recorder := &agscheduler.Recorder{
		Backend: &failingMetadataBackend{Backend: adaptertest.NewBackend()},
	}
	assert.NoError(t, scheduler.SetRecorder(recorder))

	ran := make(chan struct{}, 1)
	job := getJob()
	job.Func = runSchedulerWithMetadataFailure
	job.Args = map[string]any{"ran": ran}

	assert.NoError(t, scheduler.RunJob(job))
	select {
	case <-ran:
	case <-time.After(time.Second):
		t.Fatal("job did not run after recorder metadata failure")
	}
}

func TestSchedulerSetStoreFailureDoesNotBind(t *testing.T) {
	scheduler := &agscheduler.Scheduler{}

	err := scheduler.SetStore(&failingInitStore{Store: adaptertest.NewStore()})
	assert.Error(t, err)
	assert.Nil(t, agscheduler.GetStore(scheduler))
}

func TestSchedulerSetRecorderFailureDoesNotBind(t *testing.T) {
	scheduler := getSchedulerWithStore(t)
	recorder := &agscheduler.Recorder{
		Backend: &failingInitBackend{Backend: adaptertest.NewBackend()},
	}

	err := scheduler.SetRecorder(recorder)
	assert.Error(t, err)
	assert.False(t, scheduler.HasRecorder())
	assert.Nil(t, agscheduler.GetRecorder(scheduler))
}

func getFutureJob(fn func(context.Context, agscheduler.Job) string) agscheduler.Job {
	return agscheduler.Job{
		Name:     "FutureJob",
		Type:     agscheduler.JOB_TYPE_DATETIME,
		StartAt:  "2099-01-01 00:00:00",
		Timezone: "UTC",
		Func:     fn,
		Timeout:  "5s",
	}
}

func getJobWithoutFunc() agscheduler.Job {
	return agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "50ms",
	}
}

func getClusterNode() *agscheduler.ClusterNode {
	return &agscheduler.ClusterNode{
		EndpointMain: "127.0.0.1:36380",
		Endpoint:     "127.0.0.1:36380",
		EndpointGRPC: "127.0.0.1:36360",
		Queue:        "default",
	}
}

func getBroker() *agscheduler.Broker {
	mq := adaptertest.NewQueue()

	return &agscheduler.Broker{
		Queues: map[string]agscheduler.QueuePkg{
			"default": {
				Queue:   mq,
				Workers: 2,
			},
		},
	}
}

func getRecorder() *agscheduler.Recorder {
	mb := adaptertest.NewBackend()
	return &agscheduler.Recorder{Backend: mb}
}

func getListener() *agscheduler.Listener {
	return &agscheduler.Listener{
		Callbacks: []agscheduler.CallbackPkg{
			{
				Callback: dryCallbackScheduler,
				Event:    agscheduler.EVENT_JOB_ADDED | agscheduler.EVENT_JOB_DELETED,
			},
		},
	}
}

func TestSchedulerSetStore(t *testing.T) {
	s := &agscheduler.Scheduler{}
	assert.Nil(t, agscheduler.GetStore(s))

	store := adaptertest.NewStore()
	err := s.SetStore(store)
	assert.NoError(t, err)

	assert.NotNil(t, agscheduler.GetStore(s))
}

func TestSchedulerSetClusterNode(t *testing.T) {
	s := &agscheduler.Scheduler{}
	assert.Nil(t, agscheduler.GetClusterNode(s))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cn := getClusterNode()
	err := s.SetClusterNode(ctx, cn)
	assert.NoError(t, err)

	assert.NotNil(t, agscheduler.GetClusterNode(s))
}

func TestSchedulerSetBroker(t *testing.T) {
	s := &agscheduler.Scheduler{}
	assert.Nil(t, agscheduler.GetBroker(s))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	brk := &agscheduler.Broker{}
	err := s.SetBroker(ctx, brk)
	assert.NoError(t, err)

	assert.NotNil(t, agscheduler.GetBroker(s))
}

func TestSchedulerSetRecorder(t *testing.T) {
	s := &agscheduler.Scheduler{}
	assert.Nil(t, agscheduler.GetRecorder(s))

	rec := getRecorder()
	err := s.SetRecorder(rec)
	assert.NoError(t, err)

	assert.NotNil(t, agscheduler.GetRecorder(s))
}

func TestSchedulerSetListener(t *testing.T) {
	s := &agscheduler.Scheduler{}
	assert.Nil(t, agscheduler.GetListener(s))

	lis := getListener()
	err := s.SetListener(lis)
	assert.NoError(t, err)

	assert.NotNil(t, agscheduler.GetListener(s))
}

func TestSchedulerAddJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	defer s.Stop()
	j := getJob()
	j.Interval = "1s"
	j2 := getJob()

	j, err := s.AddJob(j)
	assert.NoError(t, err)
	_, err = s.AddJob(j2)
	assert.NoError(t, err)

	assert.Equal(t, agscheduler.JOB_STATUS_RUNNING, j.Status)

	s.Start()
	time.Sleep(500 * time.Millisecond)
}

func TestSchedulerAddJobDatetime(t *testing.T) {
	s := getSchedulerWithStore(t)
	defer s.Stop()
	j := getJob()
	j.Type = agscheduler.JOB_TYPE_DATETIME
	j.StartAt = "2023-09-22 07:30:08"

	j, err := s.AddJob(j)
	assert.NoError(t, err)

	s.Start()
	time.Sleep(50 * time.Millisecond)

	_, err = s.GetJob(j.Id)
	assert.ErrorIs(t, err, agscheduler.JobNotFoundError(j.Id))
}

func TestSchedulerAddJobUnregisteredError(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJobWithoutFunc()

	_, err := s.AddJob(j)
	assert.ErrorIs(t, err, agscheduler.FuncUnregisteredError(""))
}

func TestSchedulerAddJobTimeoutError(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()
	j.Timeout = "errorTimeout"

	_, err := s.AddJob(j)
	assert.Contains(t, err.Error(), "Timeout `"+j.Timeout+"` error")
}

func TestSchedulerRunJobPanic(t *testing.T) {
	s := getSchedulerWithStore(t)
	defer s.Stop()
	j := getJob()
	j.Func = runSchedulerPanic

	_, err := s.AddJob(j)
	assert.NoError(t, err)

	s.Start()
	time.Sleep(50 * time.Millisecond)
}

func TestSchedulerGetJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()

	assert.Empty(t, j.Id)

	j, err := s.AddJob(j)
	assert.NoError(t, err)
	j, err = s.GetJob(j.Id)
	assert.NoError(t, err)

	assert.NotEmpty(t, j.Id)
}

func TestSchedulerGetAllJobs(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()

	js, err := s.GetAllJobs()
	assert.NoError(t, err)
	assert.Len(t, js, 0)

	_, err = s.AddJob(j)
	assert.NoError(t, err)

	js, err = s.GetAllJobs()
	assert.NoError(t, err)
	assert.Len(t, js, 1)
}

func TestSchedulerUpdateJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()

	j, err := s.AddJob(j)
	assert.NoError(t, err)

	interval := "2s"
	j.Interval = interval
	j, err = s.UpdateJob(j)
	assert.NoError(t, err)

	assert.Equal(t, interval, j.Interval)
}

func TestSchedulerDeleteJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()

	j, err := s.AddJob(j)
	assert.NoError(t, err)
	err = s.DeleteJob(j.Id)
	assert.NoError(t, err)

	_, err = s.GetJob(j.Id)
	assert.ErrorIs(t, err, agscheduler.JobNotFoundError(j.Id))
}

func TestSchedulerDeleteAllJobs(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()

	_, err := s.AddJob(j)
	assert.NoError(t, err)
	err = s.DeleteAllJobs()
	assert.NoError(t, err)

	js, err := s.GetAllJobs()
	assert.NoError(t, err)
	assert.Len(t, js, 0)
}

func TestSchedulerPauseJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	defer s.Stop()
	j := getJob()

	j, err := s.AddJob(j)
	assert.NoError(t, err)

	s.Start()

	_, err = s.PauseJob(j.Id)
	assert.NoError(t, err)
	j, err = s.GetJob(j.Id)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.JOB_STATUS_PAUSED, j.Status)
}

func TestSchedulerPauseJobError(t *testing.T) {
	s := getSchedulerWithStore(t)
	_, err := s.PauseJob("1")
	assert.Error(t, err)
}

func TestSchedulerResumeJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	defer s.Stop()
	j := getJob()

	j, err := s.AddJob(j)
	assert.NoError(t, err)

	s.Start()

	_, err = s.PauseJob(j.Id)
	assert.NoError(t, err)
	j, err = s.GetJob(j.Id)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.JOB_STATUS_PAUSED, j.Status)

	_, err = s.ResumeJob(j.Id)
	assert.NoError(t, err)
	j, err = s.GetJob(j.Id)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.JOB_STATUS_RUNNING, j.Status)
}

func TestSchedulerResumeJobError(t *testing.T) {
	s := getSchedulerWithStore(t)
	_, err := s.ResumeJob("1")
	assert.Error(t, err)
}

func TestSchedulerRunJob(t *testing.T) {
	s := getSchedulerWithStore(t)
	j := getJob()

	j, err := s.AddJob(j)
	assert.NoError(t, err)

	s.Stop()

	err = s.RunJob(j)
	assert.NoError(t, err)
}

func TestSchedulerRunJobAppliesExecutionDefaults(t *testing.T) {
	preparedSchedulerJob = make(chan agscheduler.Job, 1)
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: capturePreparedSchedulerJob},
	)

	s := getSchedulerWithStore(t)
	err := s.RunJob(agscheduler.Job{
		Id:   "manual-job-id",
		Name: "Manual Job",
		Func: capturePreparedSchedulerJob,
	})
	assert.NoError(t, err)

	select {
	case j := <-preparedSchedulerJob:
		assert.Equal(t, "manual-job-id", j.Id)
		assert.Equal(t, "UTC", j.Timezone)
		assert.Equal(t, "1h", j.Timeout)
		assert.Equal(t, 1, j.MaxInstances)
		assert.NotNil(t, j.Args)
		assert.NotNil(t, j.Queues)
		assert.NotEmpty(t, j.FuncName)
	case <-time.After(time.Second):
		t.Fatal("job did not run")
	}
}

func TestSchedulerRunJobUnregisteredError(t *testing.T) {
	s := getSchedulerWithStore(t)
	err := s.RunJob(agscheduler.Job{
		Name:     "Invalid Job",
		FuncName: "test.missing",
		Timeout:  "1s",
	})
	assert.ErrorIs(t, err, agscheduler.FuncUnregisteredError("test.missing"))
}

func TestSchedulerScheduleJobUnregisteredError(t *testing.T) {
	s := getSchedulerWithStore(t)
	err := s.ScheduleJob(agscheduler.Job{
		Name:     "Invalid Job",
		FuncName: "test.missing",
		Timeout:  "1s",
	})
	assert.ErrorIs(t, err, agscheduler.FuncUnregisteredError("test.missing"))
}

func TestSchedulerBrokerSkipsUnregisteredQueuedJob(t *testing.T) {
	preparedSchedulerJob = make(chan agscheduler.Job, 1)
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: capturePreparedSchedulerJob},
	)

	s := getSchedulerWithStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	brk := getBroker()
	err := s.SetBroker(ctx, brk)
	assert.NoError(t, err)

	validJob, err := s.AddJob(getFutureJob(capturePreparedSchedulerJob))
	assert.NoError(t, err)
	validBytes, err := agscheduler.JobMarshal(validJob)
	assert.NoError(t, err)
	invalidBytes, err := agscheduler.JobMarshal(agscheduler.Job{
		Id:           "invalid-job-id",
		Name:         "Invalid Job",
		FuncName:     "test.missing",
		Timeout:      "1s",
		MaxInstances: 1,
	})
	assert.NoError(t, err)

	s.Start()
	defer func() {
		assert.NoError(t, brk.Clear("default"))
	}()
	defer s.Stop()

	queue := brk.Queues["default"].Queue
	assert.NoError(t, queue.PushJob(invalidBytes))
	assert.NoError(t, queue.PushJob(validBytes))

	select {
	case j := <-preparedSchedulerJob:
		assert.Equal(t, validJob.Id, j.Id)
	case <-time.After(time.Second):
		t.Fatal("worker did not continue after rejecting an unregistered job")
	}
}

func TestSchedulerScheduleJobLocal(t *testing.T) {
	cn := getClusterNode()
	s := getSchedulerWithStore(t)
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetClusterNode(ctx, cn)
	assert.NoError(t, err)
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	err = s.ScheduleJob(j)
	assert.NoError(t, err)
}

func TestSchedulerScheduleJobRemote(t *testing.T) {
	cn := getClusterNode()
	s := getSchedulerWithStore(t)
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetClusterNode(ctx, cn)
	assert.NoError(t, err)
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	cn.Endpoint = "test"
	err = s.ScheduleJob(j)
	assert.NoError(t, err)
}

func TestSchedulerScheduleJobClusterQueueNotExist(t *testing.T) {
	cn := getClusterNode()
	s := getSchedulerWithStore(t)
	defer s.Stop()
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetClusterNode(ctx, cn)
	assert.NoError(t, err)
	j.Queues = []string{"other"}
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	s.Start()
	time.Sleep(500 * time.Millisecond)

	err = s.ScheduleJob(j)
	assert.Error(t, err)
}

func TestSchedulerScheduleJobBrokerQueue(t *testing.T) {
	brk := getBroker()
	s := getSchedulerWithStore(t)
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetBroker(ctx, brk)
	assert.NoError(t, err)
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	err = s.ScheduleJob(j)
	assert.NoError(t, err)
}

func TestSchedulerScheduleJobBrokerQueueNotExist(t *testing.T) {
	brk := getBroker()
	brk.Queues = nil
	s := getSchedulerWithStore(t)
	j := getJob()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetBroker(ctx, brk)
	assert.NoError(t, err)
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	err = s.ScheduleJob(j)
	assert.Error(t, err)
}

func TestSchedulerBrokerGetQueues(t *testing.T) {
	brk := getBroker()
	qs := brk.GetQueues()
	assert.Len(t, qs, 1)
}

func TestSchedulerRecorderRecordMetadata(t *testing.T) {
	rec := getRecorder()
	s := getSchedulerWithStore(t)
	j := getJob()

	err := s.SetRecorder(rec)
	assert.NoError(t, err)
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	s.Start()
	time.Sleep(500 * time.Millisecond)
}

func TestSchedulerRecorderRecordMetadataTimeout(t *testing.T) {
	rec := getRecorder()
	s := getSchedulerWithStore(t)
	j := getJob()
	j.Timeout = "0.05ms"

	err := s.SetRecorder(rec)
	assert.NoError(t, err)
	_, err = s.AddJob(j)
	assert.NoError(t, err)

	s.Start()
	time.Sleep(500 * time.Millisecond)
}

func TestSchedulerStartAndStop(t *testing.T) {
	s := getSchedulerWithStore(t)
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}

func TestSchedulerRestartAfterStop(t *testing.T) {
	s := getSchedulerWithStore(t)
	s.Start()
	s.Stop()

	s.Start()
	assert.True(t, s.IsRunning())
	s.Stop()
	assert.False(t, s.IsRunning())
}

func TestSchedulerStartOnce(t *testing.T) {
	s := getSchedulerWithStore(t)
	s.Start()
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}

func TestSchedulerStopOnce(t *testing.T) {
	s := getSchedulerWithStore(t)
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}

func TestSchedulerShutdownWaitsForRunningJob(t *testing.T) {
	gracefulSchedulerStarted = make(chan struct{}, 1)
	gracefulSchedulerRelease = make(chan struct{})

	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: waitForSchedulerRelease},
	)

	s := getSchedulerWithStore(t)
	j, err := s.AddJob(getFutureJob(waitForSchedulerRelease))
	assert.NoError(t, err)

	s.Start()
	err = s.RunJob(j)
	assert.NoError(t, err)

	select {
	case <-gracefulSchedulerStarted:
	case <-time.After(time.Second):
		t.Fatal("job did not start")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- s.Shutdown(shutdownCtx)
	}()

	select {
	case err := <-shutdownDone:
		t.Fatalf("shutdown returned before the job completed: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	err = s.RunJob(j)
	assert.ErrorIs(t, err, agscheduler.ErrSchedulerShuttingDown)

	close(gracefulSchedulerRelease)

	select {
	case err := <-shutdownDone:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("shutdown did not return after the job completed")
	}

	assert.False(t, s.IsRunning())
}

func TestSchedulerShutdownCancelsJobsAfterDeadline(t *testing.T) {
	cancellationSchedulerStarted = make(chan struct{}, 1)
	cancellationSchedulerObserved = make(chan struct{})

	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: waitForSchedulerCancellation},
	)

	s := getSchedulerWithStore(t)
	j, err := s.AddJob(getFutureJob(waitForSchedulerCancellation))
	assert.NoError(t, err)

	s.Start()
	err = s.RunJob(j)
	assert.NoError(t, err)

	select {
	case <-cancellationSchedulerStarted:
	case <-time.After(time.Second):
		t.Fatal("job did not start")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err = s.Shutdown(shutdownCtx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	select {
	case <-cancellationSchedulerObserved:
	case <-time.After(time.Second):
		t.Fatal("job context was not cancelled")
	}
}

func TestSchedulerShutdownWaitsForBrokerWorker(t *testing.T) {
	gracefulSchedulerStarted = make(chan struct{}, 1)
	gracefulSchedulerRelease = make(chan struct{})

	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: waitForSchedulerRelease},
	)

	s := getSchedulerWithStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetBroker(ctx, getBroker())
	assert.NoError(t, err)

	j, err := s.AddJob(getFutureJob(waitForSchedulerRelease))
	assert.NoError(t, err)

	s.Start()
	err = s.ScheduleJob(j)
	assert.NoError(t, err)

	select {
	case <-gracefulSchedulerStarted:
	case <-time.After(time.Second):
		t.Fatal("broker job did not start")
	}

	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- s.Shutdown(context.Background())
	}()

	select {
	case err := <-shutdownDone:
		t.Fatalf("shutdown returned before the broker job completed: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	close(gracefulSchedulerRelease)

	select {
	case err := <-shutdownDone:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("shutdown did not return after the broker job completed")
	}
}

func TestCalcNextRunTimeTimezone(t *testing.T) {
	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "1s",
		Timezone: "America/New_York",
		Status:   agscheduler.JOB_STATUS_RUNNING,
	}

	nextRunTimeNew, err := agscheduler.CalcNextRunTime(j)
	assert.NoError(t, err)
	assert.Equal(t, "UTC", nextRunTimeNew.Location().String())
}

func TestCalcNextRunTime(t *testing.T) {
	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "1s",
		Timezone: "America/New_York",
		Status:   agscheduler.JOB_STATUS_RUNNING,
	}
	timezone, err := time.LoadLocation(j.Timezone)
	assert.NoError(t, err)

	j.Type = agscheduler.JOB_TYPE_DATETIME
	j.StartAt = "2023-09-22 07:30:08"
	nextRunTime, err := time.ParseInLocation(time.DateTime, j.StartAt, timezone)
	assert.NoError(t, err)
	nextRunTimeNew, err := agscheduler.CalcNextRunTime(j)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0).UTC(), nextRunTimeNew)

	j.Type = agscheduler.JOB_TYPE_INTERVAL
	interval := "1s"
	j.Interval = interval
	i, err := time.ParseDuration(interval)
	assert.NoError(t, err)
	nextRunTime = time.Now().In(timezone).Add(i)
	nextRunTimeNew, err = agscheduler.CalcNextRunTime(j)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0).UTC(), nextRunTimeNew)

	j.Type = agscheduler.JOB_TYPE_CRON
	cronExpr := "*/1 * * * *"
	j.CronExpr = cronExpr
	expr, err := cronexpr.Parse(cronExpr)
	assert.NoError(t, err)
	nextRunTime = expr.Next(time.Now().In(timezone))
	nextRunTimeNew, err = agscheduler.CalcNextRunTime(j)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(nextRunTime.Unix(), 0).UTC(), nextRunTimeNew)

	j.Status = agscheduler.JOB_STATUS_PAUSED
	nextRunTimeMax, err := agscheduler.GetNextRunTimeMax()
	assert.NoError(t, err)
	nextRunTimeNew, err = agscheduler.CalcNextRunTime(j)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(nextRunTimeMax.Unix(), 0).UTC(), nextRunTimeNew)

	j.Status = agscheduler.JOB_STATUS_RUNNING
	j.Type = "unknown"
	_, err = agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestCalcNextRunTimeTimezoneUnknown(t *testing.T) {
	j := agscheduler.Job{Timezone: "unknown"}

	_, err := agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestCalcNextRunTimeStartAtError(t *testing.T) {
	j := agscheduler.Job{
		Type:    agscheduler.JOB_TYPE_DATETIME,
		StartAt: "2023-10-22T07:30:08",
	}

	_, err := agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestCalcNextRunTimeIntervalError(t *testing.T) {
	j := agscheduler.Job{
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "2",
	}

	_, err := agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestCalcNextRunTimeCronExprError(t *testing.T) {
	j := agscheduler.Job{
		Type:     agscheduler.JOB_TYPE_CRON,
		Interval: "*/1 * * * * * * * *",
	}

	_, err := agscheduler.CalcNextRunTime(j)
	assert.Error(t, err)
}

func TestInfo(t *testing.T) {
	s := getSchedulerWithStore(t)
	brk := getBroker()
	rec := getRecorder()
	cn := getClusterNode()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.SetClusterNode(ctx, cn)
	assert.NoError(t, err)
	err = s.SetBroker(ctx, brk)
	assert.NoError(t, err)
	err = s.SetRecorder(rec)
	assert.NoError(t, err)

	info := s.Info()

	assert.Len(t, info, 5)
	assert.Equal(t, info["version"], agscheduler.Version)
}
