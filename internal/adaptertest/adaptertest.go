// Package adaptertest contains shared integration helpers for optional
// adapter modules. It deliberately depends only on the core module and the
// standard library.
package adaptertest

import (
	"context"
	"log/slog"
	"sort"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

func RunStore(t *testing.T, sto agscheduler.Store) {
	t.Helper()

	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: dryRunStore},
	)

	s := &agscheduler.Scheduler{}
	if err := s.SetStore(sto); err != nil {
		t.Fatal(err)
	}

	s.Start()
	defer s.Stop()

	j := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "1s",
		Func:     dryRunStore,
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	assert.Empty(t, j.FuncName)
	assert.Empty(t, j.Status)

	j, err := s.AddJob(j)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.JOB_STATUS_RUNNING, j.Status)
	assert.NotEmpty(t, j.FuncName)

	js, err := s.GetAllJobs()
	assert.NoError(t, err)
	assert.Len(t, js, 1)

	j.Type = agscheduler.JOB_TYPE_CRON
	j.CronExpr = "*/1 * * * *"
	j, err = s.UpdateJob(j)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.JOB_TYPE_CRON, j.Type)

	nextRunTimeMax, err := agscheduler.GetNextRunTimeMax()
	assert.NoError(t, err)

	j, err = s.PauseJob(j.Id)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.JOB_STATUS_PAUSED, j.Status)
	assert.Equal(t, nextRunTimeMax.Unix(), j.NextRunTime.Unix())

	j, err = s.ResumeJob(j.Id)
	assert.NoError(t, err)
	assert.NotEqual(t, nextRunTimeMax.Unix(), j.NextRunTime.Unix())

	assert.NoError(t, s.RunJob(j))
	assert.NoError(t, s.DeleteJob(j.Id))
	_, err = s.GetJob(j.Id)
	assert.ErrorIs(t, err, agscheduler.JobNotFoundError(j.Id))

	assert.NoError(t, s.DeleteAllJobs())
	js, err = s.GetAllJobs()
	assert.NoError(t, err)
	assert.Len(t, js, 0)
	assert.NoError(t, sto.Clear())
}

func RunQueue(t *testing.T, brk *agscheduler.Broker) {
	t.Helper()

	queueJobsCompleted.Store(0)
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: runQueuesSleep},
	)

	s := &agscheduler.Scheduler{}
	sto := &testStore{}
	assert.NoError(t, s.SetStore(sto))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	assert.NoError(t, s.SetBroker(ctx, brk))

	for i := 0; i < 6; i++ {
		job := agscheduler.Job{
			Name:    "Job" + strconv.Itoa(i+1),
			Type:    agscheduler.JOB_TYPE_DATETIME,
			StartAt: "2023-09-22 07:30:08",
			Func:    runQueuesSleep,
		}
		_, err := s.AddJob(job)
		assert.NoError(t, err)
	}

	s.Start()

	assert.Eventually(t, func() bool {
		return queueJobsCompleted.Load() == 6
	}, 30*time.Second, 250*time.Millisecond)

	for _, queueName := range sortedQueueNames(brk.Queues) {
		queueName := queueName
		assert.Eventually(t, func() bool {
			count, err := brk.CountJobs(queueName)
			return err == nil && (count == -1 || count <= 4)
		}, 10*time.Second, 250*time.Millisecond)
	}

	for _, queueName := range sortedQueueNames(brk.Queues) {
		queueName := queueName
		assert.Eventually(t, func() bool {
			count, err := brk.CountJobs(queueName)
			return err == nil && (count == -1 || count == 0)
		}, 15*time.Second, 250*time.Millisecond)
	}

	assert.NoError(t, s.DeleteAllJobs())
	s.Stop()

	for queueName := range brk.Queues {
		assert.NoError(t, brk.Clear(queueName))
	}
	assert.NoError(t, sto.Clear())
}

var queueJobsCompleted atomic.Int32

func RunBackend(t *testing.T, rec *agscheduler.Recorder) {
	t.Helper()

	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: dryRunRecorder},
	)

	s := &agscheduler.Scheduler{}
	sto := &testStore{}
	assert.NoError(t, s.SetStore(sto))
	if err := s.SetRecorder(rec); err != nil {
		t.Fatal(err)
	}

	job := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "2s",
		Func:     dryRunRecorder,
	}
	job, err := s.AddJob(job)
	assert.NoError(t, err)

	job2 := agscheduler.Job{
		Name:    "Job2",
		Type:    agscheduler.JOB_TYPE_DATETIME,
		StartAt: "2023-09-22 07:30:08",
		Func:    dryRunRecorder,
	}
	_, err = s.AddJob(job2)
	assert.NoError(t, err)

	records, total, err := rec.GetRecords(job.Id, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, records, 0)
	assert.Equal(t, int64(0), total)

	s.Start()
	defer s.Stop()

	slog.Info("Sleep 5s......")
	time.Sleep(5 * time.Second)

	records, total, err = rec.GetRecords(job.Id, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, records, 2)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, agscheduler.RECORD_STATUS_COMPLETED, records[0].Status)

	records, total, err = rec.GetRecords(job.Id, 2, 1)
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, int64(2), total)

	records, total, err = rec.GetAllRecords(1, 10)
	assert.NoError(t, err)
	assert.Len(t, records, 3)
	assert.Equal(t, int64(3), total)

	records, total, err = rec.GetAllRecords(2, 2)
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, "Job2", records[0].JobName)
	_, _, err = rec.GetAllRecords(10, 10)
	assert.NoError(t, err)

	assert.NoError(t, rec.DeleteRecords(job.Id))
	records, total, err = rec.GetAllRecords(1, 10)
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, int64(1), total)

	assert.NoError(t, rec.DeleteAllRecords())
	records, total, err = rec.GetAllRecords(1, 10)
	assert.NoError(t, err)
	assert.Len(t, records, 0)
	assert.Equal(t, int64(0), total)

	assert.NoError(t, s.DeleteAllJobs())
	assert.NoError(t, sto.Clear())
	assert.NoError(t, rec.Clear())
}

func dryRunStore(context.Context, agscheduler.Job) string { return "" }

func runQueuesSleep(context.Context, agscheduler.Job) string {
	time.Sleep(2 * time.Second)
	queueJobsCompleted.Add(1)
	return ""
}

func dryRunRecorder(context.Context, agscheduler.Job) string { return "" }

func NewStore() agscheduler.Store {
	return &testStore{}
}

func NewQueue() agscheduler.Queue {
	return &testQueue{}
}

func NewBackend() agscheduler.Backend {
	return &testBackend{}
}

type testStore struct {
	jobs []agscheduler.Job
}

func (s *testStore) Name() string         { return "test" }
func (s *testStore) Init() error          { return nil }
func (s *testStore) Clear() error         { s.jobs = nil; return nil }
func (s *testStore) DeleteAllJobs() error { s.jobs = nil; return nil }

func (s *testStore) AddJob(j agscheduler.Job) error {
	s.jobs = append(s.jobs, j)
	return nil
}

func (s *testStore) GetJob(id string) (agscheduler.Job, error) {
	for _, job := range s.jobs {
		if job.Id == id {
			return job, nil
		}
	}
	return agscheduler.Job{}, agscheduler.JobNotFoundError(id)
}

func (s *testStore) GetAllJobs() ([]agscheduler.Job, error) {
	jobs := make([]agscheduler.Job, len(s.jobs))
	copy(jobs, s.jobs)
	return jobs, nil
}

func (s *testStore) UpdateJob(j agscheduler.Job) error {
	for i := range s.jobs {
		if s.jobs[i].Id == j.Id {
			s.jobs[i] = j
			return nil
		}
	}
	return agscheduler.JobNotFoundError(j.Id)
}

func (s *testStore) DeleteJob(id string) error {
	for i := range s.jobs {
		if s.jobs[i].Id == id {
			s.jobs = append(s.jobs[:i], s.jobs[i+1:]...)
			return nil
		}
	}
	return agscheduler.JobNotFoundError(id)
}

func (s *testStore) GetNextRunTime() (time.Time, error) {
	if len(s.jobs) == 0 {
		return time.Time{}, nil
	}
	next := s.jobs[0].NextRunTime
	for _, job := range s.jobs[1:] {
		if job.NextRunTime.Before(next) {
			next = job.NextRunTime
		}
	}
	return next, nil
}

func sortedQueueNames(queues map[string]agscheduler.QueuePkg) []string {
	names := make([]string, 0, len(queues))
	for name := range queues {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type testQueue struct {
	jobs chan agscheduler.QueueMessage
}

func (q *testQueue) Name() string { return "test" }

func (q *testQueue) Init(context.Context) error {
	q.jobs = make(chan agscheduler.QueueMessage, 32)
	return nil
}

func (q *testQueue) PushJob(job []byte) error {
	q.jobs <- agscheduler.QueueMessage{Body: job}
	return nil
}

func (q *testQueue) PullJob() <-chan agscheduler.QueueMessage { return q.jobs }
func (q *testQueue) CountJobs() (int, error)                  { return len(q.jobs), nil }

func (q *testQueue) Clear() error {
	if q.jobs != nil {
		close(q.jobs)
		q.jobs = nil
	}
	return nil
}

type testBackend struct {
	records []agscheduler.Record
}

func (b *testBackend) Name() string { return "test" }
func (b *testBackend) Init() error  { return nil }
func (b *testBackend) Clear() error { b.records = nil; return nil }

func (b *testBackend) RecordMetadata(record agscheduler.Record) error {
	b.records = append(b.records, record)
	return nil
}

func (b *testBackend) RecordResult(id uint64, status, result string) error {
	for i := range b.records {
		if b.records[i].Id == id {
			b.records[i].Status = status
			b.records[i].Result = result
			b.records[i].EndAt = time.Now().UTC()
		}
	}
	return nil
}

func (b *testBackend) GetRecords(jobID string, page, pageSize int) ([]agscheduler.Record, int64, error) {
	records := make([]agscheduler.Record, 0)
	for _, record := range b.records {
		if record.JobId == jobID {
			records = append(records, record)
		}
	}
	return testPage(records, page, pageSize)
}

func (b *testBackend) GetAllRecords(page, pageSize int) ([]agscheduler.Record, int64, error) {
	records := make([]agscheduler.Record, len(b.records))
	copy(records, b.records)
	return testPage(records, page, pageSize)
}

func (b *testBackend) DeleteRecords(jobID string) error {
	kept := b.records[:0]
	for _, record := range b.records {
		if record.JobId != jobID {
			kept = append(kept, record)
		}
	}
	b.records = kept
	return nil
}

func (b *testBackend) DeleteAllRecords() error {
	b.records = nil
	return nil
}

func testPage(records []agscheduler.Record, page, pageSize int) ([]agscheduler.Record, int64, error) {
	sort.Sort(agscheduler.RecordSlice(records))
	total := int64(len(records))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	start := (page - 1) * pageSize
	if start >= len(records) {
		return []agscheduler.Record{}, total, nil
	}
	end := min(start+pageSize, len(records))
	return records[start:end], total, nil
}
