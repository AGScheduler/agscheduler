package stores

import (
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/kwkwc/agscheduler"
)

func TestGORMStore(t *testing.T) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/agscheduler?charset=utf8mb4&parseTime=True&loc=UTC"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	store := &GORMStore{DB: db}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	testAGScheduler(t, scheduler)

	store.Clean()
}
