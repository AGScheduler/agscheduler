package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/agscheduler/agscheduler"
	backends "github.com/agscheduler/agscheduler/backends/gorm"
	"github.com/agscheduler/agscheduler/internal/adaptertest"
)

func TestGormBackend(t *testing.T) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/agscheduler?charset=utf8mb4&parseTime=True&loc=UTC"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	backend := &backends.GormBackend{DB: db, TableName: "test_records"}
	recorder := &agscheduler.Recorder{Backend: backend}

	adaptertest.RunBackend(t, recorder)
}
