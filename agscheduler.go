package agscheduler

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Task struct {
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

func (t *Task) SetId() {
	t.id = strings.Replace(uuid.New().String(), "-", "", -1)
}

func (t *Task) Id() string {
	return t.id
}

type Storage interface {
	AddTask(task *Task)
	GetTaskById(id string) (*Task, error)
	GetAllTasks() []*Task
	UpdateTask(task *Task) error
	DeleteTaskById(id string) error
}

type Scheduler struct {
	Storage  Storage
	Timer    *time.Timer
	QuitChan chan struct{}
}
