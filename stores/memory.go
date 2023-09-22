package stores

import (
	"time"

	"agscheduler"
)

type MemoryStore struct {
	jobs []agscheduler.Job
}

func (s *MemoryStore) AddJob(j agscheduler.Job) {
	s.jobs = append(s.jobs, j)
}

func (s *MemoryStore) GetJob(id string) (agscheduler.Job, error) {
	for _, j := range s.jobs {
		if j.Id() == id {
			return j, nil
		}
	}
	return agscheduler.Job{}, agscheduler.JobNotFound(id)
}

func (s *MemoryStore) GetAllJobs() []agscheduler.Job {
	return s.jobs
}

func (s *MemoryStore) UpdateJob(j agscheduler.Job) error {
	for i, sj := range s.jobs {
		if sj.Id() == j.Id() {
			s.jobs[i] = j
			s.jobs[i].NextRunTime = agscheduler.CalcNextRunTime(j)
			return nil
		}
	}

	return agscheduler.JobNotFound(j.Id())
}

func (s *MemoryStore) DeleteJob(id string) error {
	for i, j := range s.jobs {
		if j.Id() == id {
			s.jobs = append(s.jobs[:i], s.jobs[i+1:]...)
			return nil
		}
	}
	return agscheduler.JobNotFound(id)
}

func (s *MemoryStore) GetNextRunTime() time.Time {
	if len(s.jobs) == 0 {
		return time.Time{}
	}

	minNextRunTime := s.jobs[0].NextRunTime
	for _, j := range s.jobs {
		if minNextRunTime.After(j.NextRunTime) {
			minNextRunTime = j.NextRunTime
		}
	}

	return minNextRunTime
}
