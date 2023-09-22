package agscheduler

import "time"

type Store interface {
	AddJob(job *Job)
	GetJob(id string) (*Job, error)
	GetAllJobs() []*Job
	UpdateJob(Job *Job) error
	DeleteJob(id string) error
	GetNextRunTime() time.Time
}
