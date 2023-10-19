package agscheduler

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/kwkwc/agscheduler/services/proto"
)

const (
	TYPE_DATETIME = "datetime"
	TYPE_INTERVAL = "interval"
	TYPE_CRON     = "cron"

	STATUS_RUNNING = "running"
	STATUS_PAUSED  = "paused"
)

type Job struct {
	Id          string         `json:"id"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	StartAt     string         `json:"start_at"`
	EndAt       string         `json:"end_at"`
	Interval    string         `json:"interval"`
	CronExpr    string         `json:"cron_expr"`
	Timezone    string         `json:"timezone"`
	Func        func(Job)      `json:"-"`
	FuncName    string         `json:"func_name"`
	Args        map[string]any `json:"args"`
	LastRunTime time.Time      `json:"last_run_time"`
	NextRunTime time.Time      `json:"next_run_time"`
	Status      string         `json:"status"`
}

func (j *Job) SetId() {
	j.Id = strings.Replace(uuid.New().String(), "-", "", -1)[:16]
}

func (j *Job) FullName() string {
	return j.Id + ":" + j.Name
}

func (j *Job) LastRunTimeWithTimezone() time.Time {
	timezone, _ := time.LoadLocation(j.Timezone)

	return j.LastRunTime.In(timezone)
}

func (j *Job) NextRunTimeWithTimezone() time.Time {
	timezone, _ := time.LoadLocation(j.Timezone)

	return j.NextRunTime.In(timezone)
}

func (j Job) String() string {
	return fmt.Sprintf(
		"Job{'Id':'%s', 'Name':'%s', 'Type':'%s', 'StartAt':'%s', 'EndAt':'%s', "+
			"'Interval':'%s', 'CronExpr':'%s', 'Timezone':'%s', "+
			"'FuncName':'%s', 'Args':'%s', "+
			"'LastRunTime':'%s', 'NextRunTime':'%s', 'Status':'%s'}",
		j.Id, j.Name, j.Type, j.StartAt, j.EndAt,
		j.Interval, j.CronExpr, j.Timezone,
		j.FuncName, j.Args,
		j.LastRunTimeWithTimezone(), j.NextRunTimeWithTimezone(), j.Status,
	)
}

func StateDump(j Job) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(j)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func StateLoad(state []byte) (Job, error) {
	var j Job
	buf := bytes.NewBuffer(state)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&j)
	if err != nil {
		return Job{}, err
	}
	return j, nil
}

func JobToPbJobPtr(j Job) *pb.Job {
	args, _ := structpb.NewStruct(j.Args)

	return &pb.Job{
		Id:          j.Id,
		Name:        j.Name,
		Type:        j.Type,
		StartAt:     j.StartAt,
		EndAt:       j.EndAt,
		Interval:    j.Interval,
		CronExpr:    j.CronExpr,
		Timezone:    j.Timezone,
		FuncName:    j.FuncName,
		Args:        args,
		LastRunTime: timestamppb.New(j.LastRunTime),
		NextRunTime: timestamppb.New(j.NextRunTime),
		Status:      j.Status,
	}
}

func PbJobPtrToJob(pbJob *pb.Job) Job {
	return Job{
		Id:          pbJob.GetId(),
		Name:        pbJob.GetName(),
		Type:        pbJob.GetType(),
		StartAt:     pbJob.GetStartAt(),
		EndAt:       pbJob.GetEndAt(),
		Interval:    pbJob.GetInterval(),
		CronExpr:    pbJob.GetCronExpr(),
		Timezone:    pbJob.GetTimezone(),
		FuncName:    pbJob.GetFuncName(),
		Args:        pbJob.GetArgs().AsMap(),
		LastRunTime: pbJob.GetLastRunTime().AsTime(),
		NextRunTime: pbJob.GetNextRunTime().AsTime(),
		Status:      pbJob.GetStatus(),
	}
}

func JobsToPbJobsPtr(js []Job) *pb.Jobs {
	pbJs := pb.Jobs{}

	for _, j := range js {
		pbJs.Jobs = append(pbJs.Jobs, JobToPbJobPtr(j))
	}

	return &pbJs
}

func PbJobsPtrToJobs(pbJs *pb.Jobs) []Job {
	js := make([]Job, 0)

	for _, pbJ := range pbJs.Jobs {
		js = append(js, PbJobPtrToJob(pbJ))
	}

	return js
}

var funcMap = make(map[string]func(Job))

func getFuncName(f func(Job)) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func RegisterFuncs(fs ...func(Job)) {
	for _, f := range fs {
		fName := getFuncName(f)
		funcMap[fName] = f
	}
}
