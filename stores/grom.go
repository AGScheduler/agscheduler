package stores

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/kwkwc/agscheduler"
)

const TABLE_NAME = "jobs"

// GORM table
type Jobs struct {
	ID          string    `gorm:"size:64;primaryKey"`
	NextRunTime time.Time `gorm:"index"`
	State       []byte    `gorm:"type:bytes;not null"`
}

// Stores jobs in a database table using GORM.
// The table will be created if it doesn't exist in the database.
type GORMStore struct {
	DB        *gorm.DB
	TableName string
}

func (s *GORMStore) Init() error {
	if s.TableName == "" {
		s.TableName = TABLE_NAME
	}

	if err := s.DB.Table(s.TableName).AutoMigrate(&Jobs{}); err != nil {
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

	return s.DB.Table(s.TableName).Create(&js).Error
}

func (s *GORMStore) GetJob(id string) (agscheduler.Job, error) {
	var js Jobs

	result := s.DB.Table(s.TableName).Where("id = ?", id).Limit(1).Find(&js)
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
	err := s.DB.Table(s.TableName).Find(&jsList).Error
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

	return s.DB.Table(s.TableName).Save(js).Error
}

func (s *GORMStore) DeleteJob(id string) error {
	return s.DB.Table(s.TableName).Where("id = ?", id).Delete(&Jobs{}).Error
}

func (s *GORMStore) DeleteAllJobs() error {
	return s.DB.Table(s.TableName).Where("1 = 1").Delete(&Jobs{}).Error
}

func (s *GORMStore) GetNextRunTime() (time.Time, error) {
	var js Jobs

	result := s.DB.Table(s.TableName).Order("next_run_time").Limit(1).Find(&js)
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
	return s.DB.Migrator().DropTable(s.TableName)
}
