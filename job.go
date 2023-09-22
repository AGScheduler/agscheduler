package agscheduler

import (
	"fmt"
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
	Func        func(Job)
	Args        []any
	LastRunTime time.Time
	NextRunTime time.Time
	Status      string // running | paused
}

func (j *Job) SetId() {
	j.id = strings.Replace(uuid.New().String(), "-", "", -1)
}

func (j *Job) Id() string {
	return j.id
}

func (j *Job) String() string {
	return fmt.Sprintf(
		"Job{'id':'%s', 'Name':'%s', 'Type':'%s', 'StartAt':'%s', 'EndAt':'%s', "+
			"'Interval':'%s', 'CronExpr':'%s', 'Args':'%s', "+
			"'LastRunTime':'%s', 'NextRunTime':'%s', 'Status':'%s'}",
		j.id, j.Name, j.Type, j.StartAt, j.EndAt,
		j.Interval, j.CronExpr, j.Args,
		j.LastRunTime, j.NextRunTime, j.Status,
	)
}
