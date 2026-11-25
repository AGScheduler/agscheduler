package agscheduler

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchedulerAcquireJobInstanceAtomic(t *testing.T) {
	s := &Scheduler{}
	key := jobInstanceKey(Job{Id: "same-job"})

	const attempts = 100
	results := make(chan bool, attempts)
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, _ := s.acquireJobInstance(key, 1)
			results <- ok
		}()
	}
	wg.Wait()
	close(results)

	acquired := 0
	for ok := range results {
		if ok {
			acquired++
		}
	}
	assert.Equal(t, 1, acquired)
	assert.Equal(t, 1, s.getJobInstanceCount(key))

	s.releaseJobInstance(key)
	assert.Equal(t, 0, s.getJobInstanceCount(key))
	ok, count := s.acquireJobInstance(key, 1)
	assert.True(t, ok)
	assert.Equal(t, 1, count)
}

func TestSchedulerJobInstancesAreIndependentByID(t *testing.T) {
	first := jobInstanceKey(Job{Id: "first-id", Name: "shared-name"})
	second := jobInstanceKey(Job{Id: "second-id", Name: "shared-name"})
	assert.NotEqual(t, first, second)

	s := &Scheduler{}
	firstOK, _ := s.acquireJobInstance(first, 1)
	secondOK, _ := s.acquireJobInstance(second, 1)
	assert.True(t, firstOK)
	assert.True(t, secondOK)
}

func TestSchedulerJobInstanceKeyFallsBackToName(t *testing.T) {
	first := jobInstanceKey(Job{Name: "same-name"})
	second := jobInstanceKey(Job{Name: "same-name"})
	assert.Equal(t, "name:same-name", first)
	assert.Equal(t, first, second)

	s := &Scheduler{}
	firstOK, _ := s.acquireJobInstance(first, 1)
	secondOK, count := s.acquireJobInstance(second, 1)
	assert.True(t, firstOK)
	assert.False(t, secondOK)
	assert.Equal(t, 1, count)
}

func TestSchedulerReleaseJobInstanceKeepsCountNonNegative(t *testing.T) {
	s := &Scheduler{}
	key := "missing"

	s.releaseJobInstance(key)
	assert.Equal(t, 0, s.getJobInstanceCount(key))
}
