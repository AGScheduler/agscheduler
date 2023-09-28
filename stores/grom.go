package stores

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/kwkwc/agscheduler"
)

type Jobs struct {
	ID          string    `gorm:"size:255;primaryKey"`
	NextRunTime time.Time `gorm:"index"`
	State       []byte    `gorm:"type:bytes;not null"`
}

func (Jobs) TableName() string {
	return "agscheduler_jobs"
}

type GORMStore struct {
	DB *gorm.DB
}

func (s *GORMStore) Init() {
	if !s.DB.Migrator().HasTable(&Jobs{}) {
		if err := s.DB.Migrator().CreateTable(&Jobs{}); err != nil {
			panic(fmt.Sprintf("Failed to create table: %s\n", err))
		}
	}
}

func (s *GORMStore) AddJob(j agscheduler.Job) error {
	state, err := s.stateDumps(j)
	if err != nil {
		return err
	}

	js := Jobs{ID: j.Id, NextRunTime: j.NextRunTime, State:state}

	return s.DB.Create(&js).Error
}

func (s *GORMStore) GetJob(id string) (agscheduler.Job, error) {
	var js Jobs

	result := s.DB.Where("id = ?", id).First(&js)
	if result.Error != nil {
		return agscheduler.Job{}, result.Error
	}
	if result.RowsAffected == 0 {
		return agscheduler.Job{}, agscheduler.JobNotFound(id)
	}

	return s.stateloads(js.State)
}

func (s *GORMStore) GetAllJobs() ([]agscheduler.Job, error) {
	var jsList []*Jobs
	err := s.DB.Find(&jsList).Error
	if err != nil {
		return nil, err
	}

	var jobList []agscheduler.Job
	for _, js := range jsList {
		aj, err := s.stateloads(js.State)
		if err != nil {
			return nil, err
		}
		jobList = append(jobList, aj)
	}

	return jobList, nil
}

func (s *GORMStore) UpdateJob(j agscheduler.Job) error {
	j.NextRunTime = agscheduler.CalcNextRunTime(j)

	state, err := s.stateDumps(j)
	if err != nil {
		return err
	}

	js := Jobs{ID: j.Id, NextRunTime: j.NextRunTime, State: state}

	return s.DB.Save(js).Error
}

func (s *GORMStore) DeleteJob(id string) error {
	return s.DB.Where("id = ?", id).Delete(&Jobs{}).Error
}

func (s *GORMStore) DeleteAllJobs() error {
	return s.DB.Where("1 = 1").Delete(&Jobs{}).Error
}

func (s *GORMStore) GetNextRunTime() (time.Time, error) {
	var js Jobs

	result := s.DB.Order("next_run_time").First(&js)
	if result.Error != nil {
		return time.Time{}, result.Error
	}
	if result.RowsAffected == 0 {
		return time.Time{}, nil
	}

	return js.NextRunTime, nil
}

func (s *GORMStore) stateDumps(j agscheduler.Job) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(j)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *GORMStore) stateloads(state []byte) (agscheduler.Job, error) {
	var j agscheduler.Job
	buf := bytes.NewBuffer(state)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&j)
	if err != nil {
		return agscheduler.Job{}, err
	}
	return j, nil
}
