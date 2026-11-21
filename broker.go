package agscheduler

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"sync"
	"time"
)

// When using a Broker, job scheduling is done in queue and no longer directly via API calls.
type Broker struct {
	// Job queues.
	// def: map[<queue>]QueuePkg
	Queues map[string]QueuePkg

	// Bind to each other and the Scheduler.
	scheduler *Scheduler

	workerCtx context.Context
	workersM  sync.Mutex

	workersRunning  bool
	workersStopping bool
	workerStop      chan struct{}
	workerDone      chan struct{}
}

type QueuePkg struct {
	Queue Queue
	// Number of workers.
	// Default: `2`
	Workers int
}

// Initialization functions for each broker,
// called when the scheduler run `SetBroker`.
func (b *Broker) init(ctx context.Context) error {
	slog.Info("Broker init...")

	b.workerCtx = ctx
	for _, qPkg := range b.Queues {
		if err := qPkg.Queue.Init(ctx); err != nil {
			return err
		}
		if qPkg.Workers <= 0 {
			qPkg.Workers = 2
		}
	}

	b.startWorkers()

	return nil
}

func (b *Broker) startWorkers() {
	b.workersM.Lock()
	defer b.workersM.Unlock()

	if b.workersRunning {
		return
	}

	slog.Info("Broker worker start.")
	ctx := b.workerCtx
	if ctx == nil {
		ctx = context.Background()
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	b.workersRunning = true
	b.workersStopping = false
	b.workerStop = stop
	b.workerDone = done

	wg := &sync.WaitGroup{}
	for _, qPkg := range b.Queues {
		for range qPkg.Workers {
			wg.Add(1)
			go func(q Queue) {
				defer wg.Done()
				b.worker(ctx, stop, q)
			}(qPkg.Queue)
		}
	}

	go func() {
		wg.Wait()
		b.workersM.Lock()
		b.workersRunning = false
		close(done)
		b.workersM.Unlock()
	}()
}

func (b *Broker) stopWorkers() {
	b.workersM.Lock()
	defer b.workersM.Unlock()

	if !b.workersRunning || b.workersStopping {
		return
	}
	b.workersStopping = true
	close(b.workerStop)
}

func (b *Broker) waitWorkers() {
	b.workersM.Lock()
	done := b.workerDone
	b.workersM.Unlock()

	if done != nil {
		<-done
	}
}

// Job worker, receiving jobs from the queue.
func (b *Broker) worker(ctx context.Context, stop <-chan struct{}, q Queue) {
	for {
		select {
		case <-stop:
			return
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-stop:
			return
		case <-ctx.Done():
			return
		case bJ, ok := <-q.PullJob():
			if !ok {
				return
			}
			j, err := JobUnmarshal(bJ)
			if err != nil {
				slog.Error(fmt.Sprintf("Job `%s` JobUnmarshal error: `%s`", bJ, err))
				continue
			}

			b.scheduler._runJobWithContext(b.scheduler.jobParentContext(), j)
		}
	}
}

// Randomly select a queue from the broker's Queues,
// if you specify a queue, filter by queue.
func (b *Broker) choiceQueue(queues []string) (string, error) {
	bqs := []string{}
	for q := range b.Queues {
		if len(queues) != 0 && !slices.Contains(queues, q) {
			continue
		}
		bqs = append(bqs, q)
	}

	bqsCount := len(bqs)
	if bqsCount != 0 {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		i := rand.Intn(bqsCount)
		return bqs[i], nil
	}

	return "", fmt.Errorf("queue not found")
}

func (b *Broker) pushJob(queue string, bJ []byte) error {
	return b.Queues[queue].Queue.PushJob(bJ)
}

// func (b *Broker) pullJob(queue string) <-chan []byte {
// 	return b.Queues[queue].Queue.PullJob()
// }

func (b *Broker) CountJobs(queue string) (int, error) {
	return b.Queues[queue].Queue.CountJobs()
}

func (b *Broker) Clear(queue string) error {
	return b.Queues[queue].Queue.Clear()
}

func (b *Broker) GetQueues() []map[string]any {
	queues := []map[string]any{}
	for qName, qPkg := range b.Queues {
		count, err := qPkg.Queue.CountJobs()
		if err != nil {
			slog.Warn(fmt.Sprintf("Broker count `%s` jobs error: %s", qName, err))
		}
		queues = append(queues, map[string]any{
			"name":    qName,
			"type":    qPkg.Queue.Name(),
			"count":   count,
			"workers": qPkg.Workers,
		})
	}

	return queues
}
