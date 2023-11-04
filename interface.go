package agscheduler

import "time"

// Defines the interface that each store must implement.
type Store interface {

	// Initialization functions for each store,
	// called when the scheduler run `SetStore`.
	Init() error

	// Add job to this store.
	AddJob(j Job) error

	// Get the job from this store.
	//  @return error `JobNotFoundError` if there are no job.
	GetJob(id string) (Job, error)

	// Get all jobs from this store.
	GetAllJobs() ([]Job, error)

	// Update job in store with a newer version.
	UpdateJob(j Job) error

	// Delete the job from this store.
	DeleteJob(id string) error

	// Delete all jobs from this store.
	DeleteAllJobs() error

	// Get the earliest next run time of all the jobs stored in this store,
	// or `time.Time{}` if there are no job.
	// Used to set the wakeup interval for the scheduler.
	GetNextRunTime() (time.Time, error)

	// Clear all resources bound to this store.
	Clear() error
}
