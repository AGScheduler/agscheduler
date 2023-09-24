package agscheduler

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	Id          string
	Name        string
	Type        string // datetime | interval | cron
	StartAt     time.Time
	EndAt       time.Time
	Interval    time.Duration
	CronExpr    string
	Timezone    string
	Func        func(Job)
	FuncName    string
	Args        []any
	LastRunTime time.Time
	NextRunTime time.Time
	Status      string // running | paused
}

func (j *Job) SetId() {
	j.Id = strings.Replace(uuid.New().String(), "-", "", -1)
}

func (j Job) String() string {
	return fmt.Sprintf(
		"Job{'id':'%s', 'Name':'%s', 'Type':'%s', 'StartAt':'%s', 'EndAt':'%s', "+
			"'Interval':'%s', 'CronExpr':'%s', 'Timezone':'%s', "+
			"'FuncName':'%s', 'Args':'%s', "+
			"'LastRunTime':'%s', 'NextRunTime':'%s', 'Status':'%s'}",
		j.Id, j.Name, j.Type, j.StartAt, j.EndAt,
		j.Interval, j.CronExpr, j.Timezone,
		j.FuncName, j.Args,
		j.LastRunTime, j.NextRunTime, j.Status,
	)
}

var funcs = make(map[string]func(Job))

func RegisterFuncs(f func(Job)) {
	fName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	funcs[fName] = f
}
