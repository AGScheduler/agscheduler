package stores

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/agscheduler/agscheduler"
)

const GORM_TABLE_NAME = "jobs"

// GORM table
type Jobs struct {
	ID          string    `gorm:"size:64;primaryKey"`
	NextRunTime time.Time `gorm:"index"`
	Data        []byte    `gorm:"type:bytes;not null"`
}

// Stores jobs in a database table using GORM.
// The table will be created if it doesn't exist in the database.
type GormStore struct {
	DB        *gorm.DB
	TableName string
}

func (s *GormStore) Name() string {
	return "GORM"
}

func (s *GormStore) Init() error {
	if s.TableName == "" {
		s.TableName = GORM_TABLE_NAME
	}

	if err := s.DB.Table(s.TableName).AutoMigrate(&Jobs{}); err != nil {
		return fmt.Errorf("failed to create table: %s", err)
	}

	return nil
}

func (s *GormStore) AddJob(j agscheduler.Job) error {
	bJ, err := agscheduler.JobMarshal(j)
	if err != nil {
		return err
	}

	js := Jobs{ID: j.Id, NextRunTime: j.NextRunTime, Data: bJ}

	return s.DB.Table(s.TableName).Create(&js).Error
}

func (s *GormStore) GetJob(id string) (agscheduler.Job, error) {
	var js Jobs

	result := s.DB.Table(s.TableName).Where("id = ?", id).Limit(1).Find(&js)
	if result.Error != nil {
		return agscheduler.Job{}, result.Error
	}
	if result.RowsAffected == 0 {
		return agscheduler.Job{}, agscheduler.JobNotFoundError(id)
	}

	return agscheduler.JobUnmarshal(js.Data)
}

func (s *GormStore) GetAllJobs() ([]agscheduler.Job, error) {
	var jsList []*Jobs
	err := s.DB.Table(s.TableName).Find(&jsList).Error
	if err != nil {
		return nil, err
	}

	var jobList []agscheduler.Job
	for _, js := range jsList {
		aj, err := agscheduler.JobUnmarshal(js.Data)
		if err != nil {
			return nil, err
		}
		jobList = append(jobList, aj)
	}

	return jobList, nil
}

func (s *GormStore) UpdateJob(j agscheduler.Job) error {
	bJ, err := agscheduler.JobMarshal(j)
	if err != nil {
		return err
	}

	js := Jobs{ID: j.Id, NextRunTime: j.NextRunTime, Data: bJ}

	return s.DB.Table(s.TableName).Save(js).Error
}

func (s *GormStore) DeleteJob(id string) error {
	return s.DB.Table(s.TableName).Where("id = ?", id).Delete(&Jobs{}).Error
}

func (s *GormStore) DeleteAllJobs() error {
	return s.DB.Table(s.TableName).Where("1 = 1").Delete(&Jobs{}).Error
}

func (s *GormStore) GetNextRunTime() (time.Time, error) {
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

func (s *GormStore) Clear() error {
	return s.DB.Migrator().DropTable(s.TableName)
}
