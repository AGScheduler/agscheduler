package agscheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/agscheduler/agscheduler/services/proto"
)

func getQueues() []map[string]any {
	return []map[string]any{
		{
			"name":    "default",
			"type":    "Memory",
			"count":   1,
			"workers": 2,
		},
		{
			"name":    "mail",
			"type":    "NSQ",
			"count":   2,
			"workers": 1,
		},
	}
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

func TestQueueToPbQueuePtr(t *testing.T) {
	qs := getQueues()
	for _, q := range qs {
		pbQ, err := QueueToPbQueuePtr(q)
		assert.NoError(t, err)

		assert.IsType(t, &pb.Queue{}, pbQ)
		assert.NotEmpty(t, pbQ)
	}
}

func TestQueuesToPbQueuesPtr(t *testing.T) {
	qs := getQueues()
	pbQs, err := QueuesToPbQueuesPtr(qs)
	assert.NoError(t, err)

	assert.IsType(t, []*pb.Queue{}, pbQs)
	assert.Len(t, pbQs, 2)
}
