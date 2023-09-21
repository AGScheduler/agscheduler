package storages

import (
	"fmt"
	"time"

	"agscheduler"
)

type MemoryStorage struct {
	Tasks []*agscheduler.Task
}

func (s *MemoryStorage) AddTask(task *agscheduler.Task) {
	s.Tasks = append(s.Tasks, task)
}

func (s *MemoryStorage) GetTaskById(id string) (*agscheduler.Task, error) {
	for _, t := range s.Tasks {
		if t.Id() == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("task with id %s not found", id)
}

func (s *MemoryStorage) GetAllTasks() []*agscheduler.Task {
	return s.Tasks
}

func (s *MemoryStorage) UpdateTask(task *agscheduler.Task) error {
	for i, t := range s.Tasks {
		if t.Id() == task.Id() {
			s.Tasks[i] = task
			s.Tasks[i].NextRunTime = time.Time{}
			return nil
		}
	}

	return fmt.Errorf("task with id %s not found", task.Id())
}

func (s *MemoryStorage) DeleteTaskById(id string) error {
	for i, t := range s.Tasks {
		if t.Id() == id {
			s.Tasks = append(s.Tasks[:i], s.Tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task with id %s not found", id)
}
