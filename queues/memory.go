package queues

import (
	"github.com/agscheduler/agscheduler"
)

// Queue jobs in an channel in RAM. Provides no persistence support.
type MemoryQueue struct {
	// Size of the queue.
	// Default: `32`
	Size int
	jobC chan []byte
}

func (q *MemoryQueue) Init() error {
	if q.Size <= 0 {
		q.Size = 32
	}

	q.jobC = make(chan []byte, q.Size)

	return nil
}

func (q *MemoryQueue) PushJob(j agscheduler.Job) error {
	bJ, err := agscheduler.StateDump(j)
	if err != nil {
		return err
	}

	q.jobC <- bJ

	return nil
}

func (q *MemoryQueue) PullJob() <-chan []byte {
	return q.jobC
}

func (q *MemoryQueue) Clear() error {
	close(q.jobC)
	return nil
}
