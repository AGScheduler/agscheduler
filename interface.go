package agscheduler

import "time"

type Store interface {
	AddJob(j Job)
	GetJob(id string) (Job, error)
	GetAllJobs() []Job
	UpdateJob(j Job) error
	DeleteJob(id string) error
	GetNextRunTime() time.Time
}
