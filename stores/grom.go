package stores

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/kwkwc/agscheduler"
)

const TABLE_NAME = "agscheduler_jobs"

type Jobs struct {
	ID          string    `gorm:"size:64;primaryKey"`
	NextRunTime time.Time `gorm:"index"`
	State       []byte    `gorm:"type:bytes;not null"`
}

type GORMStore struct {
	DB *gorm.DB
}

func (s *GORMStore) Init() error {
	if err := s.DB.Table(TABLE_NAME).AutoMigrate(&Jobs{}); err != nil {
		return fmt.Errorf("failed to create table: %s", err)
	}

	return nil
}

func (s *GORMStore) AddJob(j agscheduler.Job) error {
	state, err := agscheduler.StateDump(j)
	if err != nil {
		return err
	}

	js := Jobs{ID: j.Id, NextRunTime: j.NextRunTime, State: state}

	return s.DB.Table(TABLE_NAME).Create(&js).Error
}

func (s *GORMStore) GetJob(id string) (agscheduler.Job, error) {
	var js Jobs

	result := s.DB.Table(TABLE_NAME).Where("id = ?", id).Limit(1).Find(&js)
	if result.Error != nil {
		return agscheduler.Job{}, result.Error
	}
	if result.RowsAffected == 0 {
		return agscheduler.Job{}, agscheduler.JobNotFoundError(id)
	}

	return agscheduler.StateLoad(js.State)
}

func (s *GORMStore) GetAllJobs() ([]agscheduler.Job, error) {
	var jsList []*Jobs
	err := s.DB.Table(TABLE_NAME).Find(&jsList).Error
	if err != nil {
		return nil, err
	}

	var jobList []agscheduler.Job
	for _, js := range jsList {
		aj, err := agscheduler.StateLoad(js.State)
		if err != nil {
			return nil, err
		}
		jobList = append(jobList, aj)
	}

	return jobList, nil
}

func (s *GORMStore) UpdateJob(j agscheduler.Job) error {
	state, err := agscheduler.StateDump(j)
	if err != nil {
		return err
	}

	js := Jobs{ID: j.Id, NextRunTime: j.NextRunTime, State: state}

	return s.DB.Table(TABLE_NAME).Save(js).Error
}

func (s *GORMStore) DeleteJob(id string) error {
	return s.DB.Table(TABLE_NAME).Where("id = ?", id).Delete(&Jobs{}).Error
}

func (s *GORMStore) DeleteAllJobs() error {
	return s.DB.Table(TABLE_NAME).Where("1 = 1").Delete(&Jobs{}).Error
}

func (s *GORMStore) GetNextRunTime() (time.Time, error) {
	var js Jobs

	result := s.DB.Table(TABLE_NAME).Order("next_run_time").Limit(1).Find(&js)
	if result.Error != nil {
		return time.Time{}, result.Error
	}
	if result.RowsAffected == 0 {
		return time.Time{}, nil
	}

	nextRunTimeMin := js.NextRunTime
	return nextRunTimeMin, nil
}

func (s *GORMStore) Clear() error {
	return s.DB.Migrator().DropTable(TABLE_NAME)
}
