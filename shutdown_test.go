package agscheduler

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSchedulerWaitForShutdownSecondSignal(t *testing.T) {
	signals := make(chan os.Signal, 2)
	beforeShutdownStarted := make(chan struct{})
	releaseBeforeShutdown := make(chan struct{})
	forceExitCalled := make(chan struct{})

	result := make(chan error, 1)
	go func() {
		result <- (&Scheduler{}).waitForShutdown(
			time.Second,
			func() error {
				close(beforeShutdownStarted)
				<-releaseBeforeShutdown
				return nil
			},
			signals,
			func() {
				close(forceExitCalled)
			},
		)
	}()

	signals <- syscall.SIGINT
	<-beforeShutdownStarted
	signals <- syscall.SIGTERM

	select {
	case <-forceExitCalled:
	case <-time.After(time.Second):
		t.Fatal("second signal did not trigger force exit")
	}

	close(releaseBeforeShutdown)
	assert.NoError(t, <-result)
}
