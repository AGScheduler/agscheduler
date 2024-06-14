package agscheduler

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"time"
)

// When using a Broker, job scheduling is done in queue and no longer directly via API calls.
type Broker struct {
	// Job queues.
	// def: map[<queue>]Queue
	Queues map[string]Queue
	// Number of workers per queue.
	// Default: `2`
	WorkersPerQueue int

	// Bind to each other and the Scheduler.
	scheduler *Scheduler
}

// Initialization functions for each broker,
// called when the scheduler run `SetBroker`.
func (b *Broker) init(ctx context.Context) error {
	slog.Info("Broker init...")

	if b.WorkersPerQueue <= 0 {
		b.WorkersPerQueue = 2
	}

	slog.Info("Broker worker start.")
	for _, q := range b.Queues {
		if err := q.Init(ctx); err != nil {
			return err
		}
		for range b.WorkersPerQueue {
			go b.worker(ctx, q)
		}
	}

	return nil
}

// Job worker, receiving jobs from the queue.
func (b *Broker) worker(ctx context.Context, q Queue) {
	for {
		select {
		case <-ctx.Done():
			return
		case bJ := <-q.PullJob():
			j, err := JobUnmarshal(bJ)
			if err != nil {
				slog.Error(fmt.Sprintf("Job `%s` JobUnmarshal error: `%s`", bJ, err))
				continue
			}

			b.scheduler._runJob(j)
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

func (b *Broker) CountJobs(queue string) (int, error) {
	return b.Queues[queue].CountJobs()
}

func (b *Broker) Clear(queue string) error {
	return b.Queues[queue].Clear()
}
