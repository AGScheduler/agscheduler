package agscheduler

type Store interface {
	AddJob(job *Job)
	GetJobById(id string) (*Job, error)
	GetAllJobs() []*Job
	UpdateJob(Job *Job) error
	DeleteJobById(id string) error
}
