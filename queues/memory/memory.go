package queues

import (
	"context"

	"github.com/agscheduler/agscheduler"
)

// Queue jobs in an channel in RAM.
// Provides no persistence support.
// Cluster mode is not supported.
type MemoryQueue struct {
	// Size of the channel.
	// Default: `32`
	Size int
	jobC chan agscheduler.QueueMessage
}

func (q *MemoryQueue) Name() string {
	return "Memory"
}

func (q *MemoryQueue) Init(ctx context.Context) error {
	if q.Size <= 0 {
		q.Size = 32
	}

	q.jobC = make(chan agscheduler.QueueMessage, q.Size)

	return nil
}

func (q *MemoryQueue) PushJob(bJ []byte) error {
	q.jobC <- agscheduler.QueueMessage{Body: bJ}

	return nil
}

func (q *MemoryQueue) PullJob() <-chan agscheduler.QueueMessage {
	return q.jobC
}

func (q *MemoryQueue) CountJobs() (int, error) {
	return len(q.jobC), nil
}

func (q *MemoryQueue) Clear() error {
	defer close(q.jobC)

	return nil
}
