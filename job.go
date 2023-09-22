package agscheduler

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	id          string
	Name        string
	Type        string // datetime | interval | cron
	StartAt     time.Time
	EndAt       time.Time
	Interval    time.Duration
	CronExpr    string
	Func        func(...any)
	Args        []any
	LastRunTime time.Time
	NextRunTime time.Time
	Status      string // running | paused
}

func (t *Job) SetId() {
	t.id = strings.Replace(uuid.New().String(), "-", "", -1)
}

func (t *Job) Id() string {
	return t.id
}
