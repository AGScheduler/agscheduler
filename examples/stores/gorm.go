// go run examples/stores/base.go examples/stores/gorm.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/agscheduler/agscheduler/stores"
)

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/agscheduler?charset=utf8mb4&parseTime=True&loc=UTC"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}

	store := &stores.GormStore{DB: db, TableName: "example_jobs"}

	runExample(store)
}
