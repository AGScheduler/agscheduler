package agscheduler

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

type event uint32

// constant indicating the event.
const (
	EVENT_SCHEDULER_STARTED event = 1 << iota
	EVENT_SCHEDULER_STOPPED

	EVENT_JOB_ADDED
	EVENT_JOB_UPDATED
	EVENT_JOB_DELETED
	EVENT_ALL_JOBS_DELETED
	EVENT_JOB_PAUSED
	EVENT_JOB_RESUMED
	EVENT_JOB_EXECUTED
	EVENT_JOB_ERROR
	EVENT_JOB_TIMEOUT

	EVENT_ALL event = EVENT_SCHEDULER_STARTED | EVENT_SCHEDULER_STOPPED |
		EVENT_JOB_ADDED | EVENT_JOB_UPDATED |
		EVENT_JOB_DELETED | EVENT_ALL_JOBS_DELETED |
		EVENT_JOB_PAUSED | EVENT_JOB_RESUMED |
		EVENT_JOB_EXECUTED | EVENT_JOB_ERROR | EVENT_JOB_TIMEOUT
)

type EventPkg struct {
	Event event
	JobId string
	Data  any
}

// Event listener.
type Listener struct {
	Callbacks []CallbackPkg
}

type CallbackPkg struct {
	Callback func(ep EventPkg)
	Event    event
}

// Initialization functions for each Listener,
// called when the scheduler run `SetListener`.
func (l *Listener) init() error {
	slog.Info("Listener init...")

	return nil
}

// Event handler.
func (l *Listener) handleEvent(eP EventPkg) error {
	for _, cP := range l.Callbacks {
		if cP.Event&eP.Event == 0 {
			continue
		}

		go func(cP CallbackPkg) {
			defer func() {
				if err := recover(); err != nil {
					slog.Error(fmt.Sprintf("Listener handle event error: %s", err))
					slog.Debug(string(debug.Stack()))
				}
			}()

			cP.Callback(eP)
		}(cP)
	}

	return nil
}
