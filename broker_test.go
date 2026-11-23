package agscheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type ackTrackingQueue struct {
	deliveries chan QueueMessage
}

func (q *ackTrackingQueue) Name() string               { return "ack-tracking" }
func (q *ackTrackingQueue) Init(context.Context) error { return nil }
func (q *ackTrackingQueue) PushJob([]byte) error       { return nil }
func (q *ackTrackingQueue) PullJob() <-chan QueueMessage {
	return q.deliveries
}
func (q *ackTrackingQueue) CountJobs() (int, error) { return len(q.deliveries), nil }
func (q *ackTrackingQueue) Clear() error            { return nil }

type brokerTestStore struct {
	jobs []Job
}

func (s *brokerTestStore) Name() string { return "broker-test" }
func (s *brokerTestStore) Init() error  { return nil }
func (s *brokerTestStore) Clear() error { return s.DeleteAllJobs() }

func (s *brokerTestStore) AddJob(j Job) error {
	s.jobs = append(s.jobs, j)
	return nil
}

func (s *brokerTestStore) GetJob(id string) (Job, error) {
	for _, job := range s.jobs {
		if job.Id == id {
			return job, nil
		}
	}
	return Job{}, JobNotFoundError(id)
}

func (s *brokerTestStore) GetAllJobs() ([]Job, error) {
	jobs := make([]Job, len(s.jobs))
	copy(jobs, s.jobs)
	return jobs, nil
}

func (s *brokerTestStore) UpdateJob(j Job) error {
	for i := range s.jobs {
		if s.jobs[i].Id == j.Id {
			s.jobs[i] = j
			return nil
		}
	}
	return JobNotFoundError(j.Id)
}

func (s *brokerTestStore) DeleteJob(id string) error {
	for i := range s.jobs {
		if s.jobs[i].Id == id {
			s.jobs = append(s.jobs[:i], s.jobs[i+1:]...)
			return nil
		}
	}
	return JobNotFoundError(id)
}

func (s *brokerTestStore) DeleteAllJobs() error {
	s.jobs = nil
	return nil
}

func (s *brokerTestStore) GetNextRunTime() (time.Time, error) {
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

func preserveFuncMap(t *testing.T) {
	t.Helper()

	original := make(map[string]FuncPkg, len(FuncMap))
	for name, pkg := range FuncMap {
		original[name] = pkg
	}
	t.Cleanup(func() {
		FuncMap = original
	})
}

var (
	brokerAckStarted   chan struct{}
	brokerAckRelease   chan struct{}
	brokerNackStarted  chan struct{}
	brokerNackCanceled chan struct{}
)

func waitForBrokerAck(ctx context.Context, j Job) string {
	brokerAckStarted <- struct{}{}
	<-brokerAckRelease
	return ""
}

func waitForBrokerNack(ctx context.Context, j Job) string {
	brokerNackStarted <- struct{}{}
	<-ctx.Done()
	close(brokerNackCanceled)
	return ""
}

func getBroker() *Broker {
	return &Broker{
		Queues: map[string]QueuePkg{
			"default": {
				Queue:   nil,
				Workers: 2,
			},
		},
	}
}

func TestChoiceQueue(t *testing.T) {
	brk := getBroker()

	queue, err := brk.choiceQueue([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "default", queue)
}

func TestChoiceQueueFilter(t *testing.T) {
	brk := getBroker()

	queue, err := brk.choiceQueue([]string{"mail"})
	assert.Error(t, err)
	assert.Equal(t, "", queue)
}

func TestBrokerAcknowledgesAfterJobCompletion(t *testing.T) {
	preserveFuncMap(t)
	brokerAckStarted = make(chan struct{}, 1)
	brokerAckRelease = make(chan struct{})
	RegisterFuncs(FuncPkg{Func: waitForBrokerAck})

	scheduler := &Scheduler{}
	assert.NoError(t, scheduler.SetStore(&brokerTestStore{}))

	queue := &ackTrackingQueue{deliveries: make(chan QueueMessage, 1)}
	broker := &Broker{Queues: map[string]QueuePkg{
		"default": {Queue: queue, Workers: 1},
	}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	assert.NoError(t, scheduler.SetBroker(ctx, broker))

	job, err := scheduler.AddJob(Job{
		Name:    "Ack Job",
		Type:    JOB_TYPE_DATETIME,
		StartAt: "2099-01-01 00:00:00",
		Func:    waitForBrokerAck,
		Timeout: "5s",
	})
	assert.NoError(t, err)
	body, err := JobMarshal(job)
	assert.NoError(t, err)

	acked := make(chan struct{}, 1)
	nacked := make(chan struct{}, 1)
	queue.deliveries <- QueueMessage{
		Body: body,
		Ack: func() error {
			acked <- struct{}{}
			return nil
		},
		Nack: func() error {
			nacked <- struct{}{}
			return nil
		},
	}

	scheduler.Start()
	defer scheduler.Stop()

	select {
	case <-brokerAckStarted:
	case <-time.After(time.Second):
		t.Fatal("job did not start")
	}
	select {
	case <-acked:
		t.Fatal("message was acknowledged before the job completed")
	default:
	}

	close(brokerAckRelease)
	select {
	case <-acked:
	case <-time.After(time.Second):
		t.Fatal("message was not acknowledged after the job completed")
	}
	select {
	case <-nacked:
		t.Fatal("message was rejected after a successful run")
	default:
	}
}

func TestBrokerNacksMessageInterruptedByShutdown(t *testing.T) {
	preserveFuncMap(t)
	brokerNackStarted = make(chan struct{}, 1)
	brokerNackCanceled = make(chan struct{})
	RegisterFuncs(FuncPkg{Func: waitForBrokerNack})

	scheduler := &Scheduler{}
	assert.NoError(t, scheduler.SetStore(&brokerTestStore{}))

	queue := &ackTrackingQueue{deliveries: make(chan QueueMessage, 1)}
	broker := &Broker{Queues: map[string]QueuePkg{
		"default": {Queue: queue, Workers: 1},
	}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	assert.NoError(t, scheduler.SetBroker(ctx, broker))

	job, err := scheduler.AddJob(Job{
		Name:    "Nack Job",
		Type:    JOB_TYPE_DATETIME,
		StartAt: "2099-01-01 00:00:00",
		Func:    waitForBrokerNack,
		Timeout: "5s",
	})
	assert.NoError(t, err)
	body, err := JobMarshal(job)
	assert.NoError(t, err)

	acked := make(chan struct{}, 1)
	nacked := make(chan struct{}, 1)
	queue.deliveries <- QueueMessage{
		Body: body,
		Ack: func() error {
			acked <- struct{}{}
			return nil
		},
		Nack: func() error {
			nacked <- struct{}{}
			return nil
		},
	}

	scheduler.Start()
	select {
	case <-brokerNackStarted:
	case <-time.After(time.Second):
		t.Fatal("job did not start")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer shutdownCancel()
	err = scheduler.Shutdown(shutdownCtx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	select {
	case <-nacked:
	case <-time.After(time.Second):
		t.Fatal("interrupted message was not returned to the queue")
	}
	select {
	case <-acked:
		t.Fatal("interrupted message was acknowledged")
	default:
	}
}
