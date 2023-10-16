package stores

import (
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

func (s *GORMStore) Init() error {
	if !s.DB.Migrator().HasTable(&Jobs{}) {
		if err := s.DB.Migrator().CreateTable(&Jobs{}); err != nil {
			return fmt.Errorf("failed to create table: %s", err)
		}
	}

	return nil
}

func (s *GORMStore) AddJob(j agscheduler.Job) error {
	state, err := agscheduler.StateDumps(j)
	if err != nil {
		return err
	}

	js := Jobs{ID: j.Id, NextRunTime: j.NextRunTime, State: state}

	return s.DB.Create(&js).Error
}

func (s *GORMStore) GetJob(id string) (agscheduler.Job, error) {
	var js Jobs

	result := s.DB.Where("id = ?", id).Limit(1).Find(&js)
	if result.Error != nil {
		return agscheduler.Job{}, result.Error
	}
	if result.RowsAffected == 0 {
		return agscheduler.Job{}, agscheduler.JobNotFoundError(id)
	}

	return agscheduler.StateLoads(js.State)
}

func (s *GORMStore) GetAllJobs() ([]agscheduler.Job, error) {
	var jsList []*Jobs
	err := s.DB.Find(&jsList).Error
	if err != nil {
		return nil, err
	}

	var jobList []agscheduler.Job
	for _, js := range jsList {
		aj, err := agscheduler.StateLoads(js.State)
		if err != nil {
			return nil, err
		}
		jobList = append(jobList, aj)
	}

	return jobList, nil
}

func (s *GORMStore) UpdateJob(j agscheduler.Job) error {
	nextRunTime, err := agscheduler.CalcNextRunTime(j)
	if err != nil {
		return err
	}
	j.NextRunTime = nextRunTime

	state, err := agscheduler.StateDumps(j)
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

	result := s.DB.Order("next_run_time").Limit(1).Find(&js)
	if result.Error != nil {
		return time.Time{}, result.Error
	}
	if result.RowsAffected == 0 {
		return time.Time{}, nil
	}

	nextRunTimeMin := js.NextRunTime
	return nextRunTimeMin, nil
}
