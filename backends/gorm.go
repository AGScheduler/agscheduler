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
	Result  string    `gorm:"type:text"`
	StartAt time.Time `gorm:"not null"`
	EndAt   time.Time `gorm:"default:null"`
}

// Store job records in a database table using GORM.
// The table will be created if it doesn't exist in the database.
type GormBackend struct {
	DB        *gorm.DB
	TableName string
}

func (b *GormBackend) Name() string {
	return "GORM"
}

func (b *GormBackend) Init() error {
	if b.TableName == "" {
		b.TableName = GORM_TABLE_NAME
	}

	if err := b.DB.Table(b.TableName).AutoMigrate(&Records{}); err != nil {
		return fmt.Errorf("failed to create table: %s", err)
	}

	return nil
}

func (b *GormBackend) RecordMetadata(r agscheduler.Record) error {
	rs := Records{
		ID:      r.Id,
		JobId:   r.JobId,
		JobName: r.JobName,
		Status:  r.Status,
		Result:  r.Result,
		StartAt: r.StartAt,
		EndAt:   r.EndAt,
	}

	return b.DB.Table(b.TableName).Create(&rs).Error
}

func (b *GormBackend) RecordResult(id uint64, status string, result string) error {
	return b.DB.Table(b.TableName).Where("id = ?", id).
		Update("status", status).
		Update("result", result).
		Update("end_at", time.Now().UTC()).
		Error
}

func (b *GormBackend) _getRecords(page, pageSize int, query any, args ...any) ([]agscheduler.Record, int64, error) {
	var rsList []*Records
	total := int64(0)

	err := b.DB.Table(b.TableName).Where(query, args...).
		Order("start_at desc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&rsList).Error
	if err != nil {
		return nil, total, err
	}

	err = b.DB.Table(b.TableName).Where(query, args...).
		Count(&total).Error
	if err != nil {
		return nil, total, err
	}

	recordList := []agscheduler.Record{}
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

	return recordList, total, nil
}

func (b *GormBackend) GetRecords(jId string, page, pageSize int) ([]agscheduler.Record, int64, error) {
	return b._getRecords(page, pageSize, "job_id = ?", jId)
}

func (b *GormBackend) GetAllRecords(page, pageSize int) ([]agscheduler.Record, int64, error) {
	return b._getRecords(page, pageSize, "1 = 1")
}

func (b *GormBackend) DeleteRecords(jId string) error {
	return b.DB.Table(b.TableName).Where("job_id = ?", jId).Delete(&Records{}).Error
}

func (b *GormBackend) DeleteAllRecords() error {
	return b.DB.Table(b.TableName).Where("1 = 1").Delete(&Records{}).Error
}

func (b *GormBackend) Clear() error {
	return b.DB.Migrator().DropTable(b.TableName)
}
