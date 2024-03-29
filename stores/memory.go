package stores

import (
	"sort"
	"time"

	"github.com/agscheduler/agscheduler"
)

// Stores jobs in an array in RAM. Provides no persistence support.
type MemoryStore struct {
	jobs []agscheduler.Job
}

func (s *MemoryStore) Init() error {
	return nil
}

func (s *MemoryStore) AddJob(j agscheduler.Job) error {
	s.jobs = append(s.jobs, j)
	return nil
}

func (s *MemoryStore) GetJob(id string) (agscheduler.Job, error) {
	for _, j := range s.jobs {
		if j.Id == id {
			return j, nil
		}
	}
	return agscheduler.Job{}, agscheduler.JobNotFoundError(id)
}

func (s *MemoryStore) GetAllJobs() ([]agscheduler.Job, error) {
	return s.jobs, nil
}

func (s *MemoryStore) UpdateJob(j agscheduler.Job) error {
	for i, sj := range s.jobs {
		if sj.Id == j.Id {
			s.jobs[i] = j

			return nil
		}
	}

	return agscheduler.JobNotFoundError(j.Id)
}

func (s *MemoryStore) DeleteJob(id string) error {
	for i, j := range s.jobs {
		if j.Id == id {
			s.jobs = append(s.jobs[:i], s.jobs[i+1:]...)
			return nil
		}
	}
	return agscheduler.JobNotFoundError(id)
}

func (s *MemoryStore) DeleteAllJobs() error {
	s.jobs = nil
	return nil
}

func (s *MemoryStore) GetNextRunTime() (time.Time, error) {
	if len(s.jobs) == 0 {
		return time.Time{}, nil
	}

	js := make([]agscheduler.Job, len(s.jobs))
	copy(js, s.jobs)
	sort.Sort(agscheduler.JobSlice(js))

	nextRunTimeMin := js[0].NextRunTime
	return nextRunTimeMin, nil
}

func (s *MemoryStore) Clear() error {
	return s.DeleteAllJobs()
}
