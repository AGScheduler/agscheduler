package agscheduler

import "time"

type Store interface {
	Init() error
	AddJob(j Job) error
	GetJob(id string) (Job, error)
	GetAllJobs() ([]Job, error)
	UpdateJob(j Job) error
	DeleteJob(id string) error
	DeleteAllJobs() error
	GetNextRunTime() (time.Time, error)
}
