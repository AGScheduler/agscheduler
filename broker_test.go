package agscheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
