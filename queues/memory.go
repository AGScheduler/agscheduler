package queues

import "context"

// Queue jobs in an channel in RAM.
// Provides no persistence support.
// Cluster mode is not supported.
type MemoryQueue struct {
	// Size of the channel.
	// Default: `32`
	Size int
	jobC chan []byte
}

func (q *MemoryQueue) Init(ctx context.Context) error {
	if q.Size <= 0 {
		q.Size = 32
	}

	q.jobC = make(chan []byte, q.Size)

	return nil
}

func (q *MemoryQueue) PushJob(bJ []byte) error {
	q.jobC <- bJ

	return nil
}

func (q *MemoryQueue) PullJob() <-chan []byte {
	return q.jobC
}

func (q *MemoryQueue) Clear() error {
	defer close(q.jobC)

	return nil
}
