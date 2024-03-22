package agscheduler

import (
	"bytes"
	"context"
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

// constant indicating a job's type
const (
	TYPE_DATETIME = "datetime"
	TYPE_INTERVAL = "interval"
	TYPE_CRON     = "cron"
)

// constant indicating a job's status
const (
	STATUS_RUNNING = "running"
	STATUS_PAUSED  = "paused"
)

// Carry the information of the scheduled job
type Job struct {
	// The unique identifier of this job, automatically generated.
	// It should not be set manually.
	Id string `json:"id"`
	// User defined.
	Name string `json:"name"`
	// Optional: `TYPE_DATETIME` | `TYPE_INTERVAL` | `TYPE_CRON`
	Type string `json:"type"`
	// It can be used when Type is `TYPE_DATETIME`.
	StartAt string `json:"start_at"`
	// This field is useless.
	EndAt string `json:"end_at"`
	// It can be used when Type is `TYPE_INTERVAL`.
	Interval string `json:"interval"`
	// It can be used when Type is `TYPE_CRON`.
	CronExpr string `json:"cron_expr"`
	// Refer to `time.LoadLocation`.
	// Default: `UTC`
	Timezone string `json:"timezone"`
	// The job actually runs the function,
	// and you need to register it through 'RegisterFuncs' before using it.
	// Since it cannot be stored by serialization,
	// when using gRPC or HTTP calls, you should use `FuncName`.
	Func func(context.Context, Job) `json:"-"`
	// The actual path of `Func`.
	// This field has a higher priority than `Func`
	FuncName string `json:"func_name"`
	// Arguments for `Func`.
	Args map[string]any `json:"args"`
	// The running timeout of `Func`.
	// Default: `1h`
	Timeout string `json:"timeout"`
	// Used in cluster mode, if empty, randomly pick a node to run `Func`.
	Queues []string `json:"queues"`

	// Automatic update, not manual setting.
	LastRunTime time.Time `json:"last_run_time"`
	// Automatic update, not manual setting.
	// When the job is paused, this field is set to `9999-09-09 09:09:09`.
	NextRunTime time.Time `json:"next_run_time"`
	// Optional: `STATUS_RUNNING` | `STATUS_PAUSED`
	// It should not be set manually.
	Status string `json:"status"`
}

// `sort.Interface`, sorted by 'NextRunTime', ascend.
type JobSlice []Job

func (js JobSlice) Len() int           { return len(js) }
func (js JobSlice) Less(i, j int) bool { return js[i].NextRunTime.Before(js[j].NextRunTime) }
func (js JobSlice) Swap(i, j int)      { js[i], js[j] = js[j], js[i] }

func (j *Job) setId() {
	j.Id = strings.Replace(uuid.New().String(), "-", "", -1)[:16]
}

// Initialization functions for each job,
// called when the scheduler run `AddJob`.
func (j *Job) init() error {
	j.setId()

	j.Status = STATUS_RUNNING

	if j.Timezone == "" {
		j.Timezone = "UTC"
	}

	if j.FuncName == "" {
		j.FuncName = getFuncName(j.Func)
	}

	if j.Timeout == "" {
		j.Timeout = "1h"
	}

	nextRunTime, err := CalcNextRunTime(*j)
	if err != nil {
		return err
	}
	j.NextRunTime = nextRunTime

	if err := j.check(); err != nil {
		return err
	}

	return nil
}

// Called when the job run `init` or scheduler run `UpdateJob`.
func (j *Job) check() error {
	if _, ok := funcMap[j.FuncName]; !ok {
		return FuncUnregisteredError(j.FuncName)
	}

	_, err := time.ParseDuration(j.Timeout)
	if err != nil {
		return &JobTimeoutError{FullName: j.FullName(), Timeout: j.Timeout, Err: err}
	}

	return nil
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

func GetNextRunTimeMax() (time.Time, error) {
	return time.ParseInLocation(
		time.DateTime,
		"9999-09-09 09:09:09",
		time.Now().UTC().Location(),
	)
}

func (j Job) String() string {
	return fmt.Sprintf(
		"Job{'Id':'%s', 'Name':'%s', 'Type':'%s', 'StartAt':'%s', 'EndAt':'%s', "+
			"'Interval':'%s', 'CronExpr':'%s', 'Timezone':'%s', "+
			"'FuncName':'%s', 'Args':'%s', 'Timeout':'%s', 'Queues':'%s', "+
			"'LastRunTime':'%s', 'NextRunTime':'%s', 'Status':'%s'}",
		j.Id, j.Name, j.Type, j.StartAt, j.EndAt,
		j.Interval, j.CronExpr, j.Timezone,
		j.FuncName, j.Args, j.Timeout, j.Queues,
		j.LastRunTimeWithTimezone(), j.NextRunTimeWithTimezone(), j.Status,
	)
}

// Serialize Job and convert to Bytes
func StateDump(j Job) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(j)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize Bytes and convert to Job
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

// Used to gRPC Protobuf
func JobToPbJobPtr(j Job) (*pb.Job, error) {
	args, err := structpb.NewStruct(j.Args)
	if err != nil {
		return &pb.Job{}, err
	}

	pbJ := &pb.Job{
		Id:       j.Id,
		Name:     j.Name,
		Type:     j.Type,
		StartAt:  j.StartAt,
		EndAt:    j.EndAt,
		Interval: j.Interval,
		CronExpr: j.CronExpr,
		Timezone: j.Timezone,
		FuncName: j.FuncName,
		Args:     args,
		Timeout:  j.Timeout,
		Queues:   j.Queues,

		LastRunTime: timestamppb.New(j.LastRunTime),
		NextRunTime: timestamppb.New(j.NextRunTime),
		Status:      j.Status,
	}

	return pbJ, nil
}

// Used to gRPC Protobuf
func PbJobPtrToJob(pbJob *pb.Job) Job {
	return Job{
		Id:       pbJob.GetId(),
		Name:     pbJob.GetName(),
		Type:     pbJob.GetType(),
		StartAt:  pbJob.GetStartAt(),
		EndAt:    pbJob.GetEndAt(),
		Interval: pbJob.GetInterval(),
		CronExpr: pbJob.GetCronExpr(),
		Timezone: pbJob.GetTimezone(),
		FuncName: pbJob.GetFuncName(),
		Args:     pbJob.GetArgs().AsMap(),
		Timeout:  pbJob.GetTimeout(),
		Queues:   pbJob.GetQueues(),

		LastRunTime: pbJob.GetLastRunTime().AsTime(),
		NextRunTime: pbJob.GetNextRunTime().AsTime(),
		Status:      pbJob.GetStatus(),
	}
}

// Used to gRPC Protobuf
func JobsToPbJobsPtr(js []Job) (*pb.Jobs, error) {
	pbJs := &pb.Jobs{}

	for _, j := range js {
		pbJ, err := JobToPbJobPtr(j)
		if err != nil {
			return &pb.Jobs{}, err
		}

		pbJs.Jobs = append(pbJs.Jobs, pbJ)
	}

	return pbJs, nil
}

// Used to gRPC Protobuf
func PbJobsPtrToJobs(pbJs *pb.Jobs) []Job {
	js := make([]Job, 0)

	for _, pbJ := range pbJs.Jobs {
		js = append(js, PbJobPtrToJob(pbJ))
	}

	return js
}

type FuncPkg struct {
	Func func(context.Context, Job)
	// About this function.
	Info string
}

// Record the actual path of function and the corresponding function.
// Since golang can't serialize functions,
// need to register them with `RegisterFuncs` before using it.
var funcMap = make(map[string]FuncPkg)

func getFuncName(f func(context.Context, Job)) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func RegisterFuncs(fps ...FuncPkg) {
	for _, fp := range fps {
		fName := getFuncName(fp.Func)
		funcMap[fName] = fp
	}
}
