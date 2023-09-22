package stores

import (
	"time"

	"agscheduler"
)

type MemoryStore struct {
	Jobs []*agscheduler.Job
}

func (s *MemoryStore) AddJob(job *agscheduler.Job) {
	s.Jobs = append(s.Jobs, job)
}

func (s *MemoryStore) GetJob(id string) (*agscheduler.Job, error) {
	for _, j := range s.Jobs {
		if j.Id() == id {
			return j, nil
		}
	}
	return nil, agscheduler.JobNotFound(id)
}

func (s *MemoryStore) GetAllJobs() []*agscheduler.Job {
	return s.Jobs
}

func (s *MemoryStore) UpdateJob(job *agscheduler.Job) error {
	for i, j := range s.Jobs {
		if j.Id() == job.Id() {
			s.Jobs[i] = job
			s.Jobs[i].NextRunTime = agscheduler.CalcNextRunTime(job)
			return nil
		}
	}

	return agscheduler.JobNotFound(job.Id())
}

func (s *MemoryStore) DeleteJob(id string) error {
	for i, j := range s.Jobs {
		if j.Id() == id {
			s.Jobs = append(s.Jobs[:i], s.Jobs[i+1:]...)
			return nil
		}
	}
	return agscheduler.JobNotFound(id)
}

func (s *MemoryStore) GetNextRunTime() time.Time {
	if len(s.Jobs) == 0 {
		return time.Time{}
	}

	minNextRunTime := s.Jobs[0].NextRunTime
	for _, j := range s.Jobs {
		if minNextRunTime.After(j.NextRunTime) {
			minNextRunTime = j.NextRunTime
		}
	}

	return minNextRunTime
}
