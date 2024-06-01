package backends

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/agscheduler/agscheduler"
)

const GORM_TABLE_NAME = "records"

// GORM table
type Records struct {
	ID      uint64    `gorm:"primaryKey"`
	JobId   string    `gorm:"size:64;not null"`
	JobName string    `gorm:"size:64"`
	Status  string    `gorm:"size:9;not null"`
	Result  []byte    `gorm:"type:bytes"`
	StartAt time.Time `gorm:"not null"`
	EndAt   time.Time `gorm:"default:null"`
}

// Store job records in a database table using GORM.
// The table will be created if it doesn't exist in the database.
type GORMBackend struct {
	DB        *gorm.DB
	TableName string
}

func (b *GORMBackend) Init() error {
	if b.TableName == "" {
		b.TableName = GORM_TABLE_NAME
	}

	if err := b.DB.Table(b.TableName).AutoMigrate(&Records{}); err != nil {
		return fmt.Errorf("failed to create table: %s", err)
	}

	return nil
}

func (b *GORMBackend) RecordMetadata(r agscheduler.Record) error {
	rs := Records{
		ID:      r.Id,
		JobId:   r.JobId,
		JobName: r.JobName,
		Status:  r.Status,
		StartAt: r.StartAt,
	}

	return b.DB.Table(b.TableName).Create(&rs).Error
}

func (b *GORMBackend) RecordResult(id uint64, status string, result []byte) error {
	return b.DB.Table(b.TableName).Where("id = ?", id).
		Update("status", status).
		Update("result", result).
		Update("end_at", time.Now().UTC()).
		Error
}

func (b *GORMBackend) GetRecords(jId string) ([]agscheduler.Record, error) {
	var rsList []*Records

	err := b.DB.Table(b.TableName).Where("job_id = ?", jId).
		Order("start_at desc").
		Find(&rsList).Error
	if err != nil {
		return nil, err
	}

	var recordList []agscheduler.Record
	for _, rs := range rsList {
		recordList = append(recordList, agscheduler.Record{
			Id:      rs.ID,
			JobId:   rs.JobId,
			JobName: rs.JobName,
			Status:  rs.Status,
			Result:  rs.Result,
			StartAt: rs.StartAt,
			EndAt:   rs.EndAt,
		})
	}

	return recordList, nil
}

func (b *GORMBackend) GetAllRecords() ([]agscheduler.Record, error) {
	var rsList []*Records

	err := b.DB.Table(b.TableName).Order("start_at desc").Find(&rsList).Error
	if err != nil {
		return nil, err
	}

	var recordList []agscheduler.Record
	for _, rs := range rsList {
		recordList = append(recordList, agscheduler.Record{
			Id:      rs.ID,
			JobId:   rs.JobId,
			JobName: rs.JobName,
			Status:  rs.Status,
			Result:  rs.Result,
			StartAt: rs.StartAt,
			EndAt:   rs.EndAt,
		})
	}

	return recordList, nil
}

func (b *GORMBackend) DeleteRecords(jId string) error {
	return b.DB.Table(b.TableName).Where("job_id = ?", jId).Delete(&Records{}).Error
}

func (b *GORMBackend) DeleteAllRecords() error {
	return b.DB.Table(b.TableName).Where("1 = 1").Delete(&Records{}).Error
}

func (b *GORMBackend) Clear() error {
	return b.DB.Migrator().DropTable(b.TableName)
}
