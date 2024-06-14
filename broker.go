package agscheduler

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"time"

	pb "github.com/agscheduler/agscheduler/services/proto"
)

// When using a Broker, job scheduling is done in queue and no longer directly via API calls.
type Broker struct {
	// Job queues.
	// def: map[<queue>]QueuePkg
	Queues map[string]QueuePkg

	// Bind to each other and the Scheduler.
	scheduler *Scheduler
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

	slog.Info("Broker worker start.")
	for _, qPkg := range b.Queues {
		if err := qPkg.Queue.Init(ctx); err != nil {
			return err
		}
		if qPkg.Workers <= 0 {
			qPkg.Workers = 2
		}
		for range qPkg.Workers {
			go b.worker(ctx, qPkg.Queue)
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

func (b *Broker) pushJob(queue string, bJ []byte) error {
	return b.Queues[queue].Queue.PushJob(bJ)
}

func (b *Broker) pullJob(queue string) <-chan []byte {
	return b.Queues[queue].Queue.PullJob()
}

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

// Used to gRPC Protobuf
func QueueToPbQueuePtr(q map[string]any) (*pb.Queue, error) {
	pbQ := &pb.Queue{
		Name:    q["name"].(string),
		Type:    q["type"].(string),
		Count:   int64(q["count"].(int)),
		Workers: int32(q["workers"].(int)),
	}

	return pbQ, nil
}

// Used to gRPC Protobuf
func QueuesToPbQueuesPtr(qs []map[string]any) ([]*pb.Queue, error) {
	pbQs := []*pb.Queue{}

	for _, q := range qs {
		pbQ, err := QueueToPbQueuePtr(q)
		if err != nil {
			return []*pb.Queue{}, err
		}

		pbQs = append(pbQs, pbQ)
	}

	return pbQs, nil
}
