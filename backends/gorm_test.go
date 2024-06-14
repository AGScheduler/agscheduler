package backends

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/agscheduler/agscheduler"
)

func TestGormBackend(t *testing.T) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/agscheduler?charset=utf8mb4&parseTime=True&loc=UTC"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	gb := &GormBackend{DB: db, TableName: "test_records"}
	recorder := &agscheduler.Recorder{Backend: gb}

	runTest(t, recorder)
}
