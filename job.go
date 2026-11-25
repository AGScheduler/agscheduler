package agscheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// constant indicating a job's type
const (
	JOB_TYPE_DATETIME = "datetime"
	JOB_TYPE_INTERVAL = "interval"
	JOB_TYPE_CRON     = "cron"
)

// constant indicating a job's status
const (
	JOB_STATUS_RUNNING = "running"
	JOB_STATUS_PAUSED  = "paused"
)

// Carry the information of the scheduled job
type Job struct {
	// The unique identifier of this job, automatically generated.
	// It should not be set manually.
	Id string `json:"id"`
	// User defined.
	Name string `json:"name"`
	// Optional: `JOB_TYPE_DATETIME` | `JOB_TYPE_INTERVAL` | `JOB_TYPE_CRON`
	Type string `json:"type"`
	// It can be used when Type is `JOB_TYPE_DATETIME`.
	// e.g. `2023-09-22 07:30:08`
	StartAt string `json:"start_at"`
	// This field is useless.
	EndAt string `json:"end_at"`
	// It can be used when Type is `JOB_TYPE_INTERVAL`.
	// Must be a positive whole number of seconds.
	// e.g. `2s`
	Interval string `json:"interval"`
	// It can be used when Type is `JOB_TYPE_CRON`.
	// See `https://en.wikipedia.org/wiki/Cron`.
	// e.g. `*/1 * * * *`
	CronExpr string `json:"cron_expr"`
	// Refer to `time.LoadLocation`.
	// See `https://en.wikipedia.org/wiki/List_of_tz_database_time_zones`
	// Default: `UTC`
	Timezone string `json:"timezone"`
	// The job actually runs the function,
	// and you need to register it through 'RegisterFuncs' before using it.
	// Since it cannot be stored by serialization,
	// when using gRPC or HTTP calls, you should use `FuncName`.
	Func func(context.Context, Job) (result string) `json:"-"`
	// The actual path of `Func`.
	// This field has a higher priority than `Func`
	// e.g. `main.xxxFunc`
	//      `github.com/agscheduler/agscheduler/examples.PrintMsg`
	FuncName string `json:"func_name"`
	// Arguments for `Func`.
	Args map[string]any `json:"args"`
	// The running timeout of `Func`.
	// Must be greater than zero.
	// Default: `1h`
	Timeout string `json:"timeout"`
	// Used in cluster mode, if empty, randomly pick a node to run `Func`,
	// but when broker exist, if empty, randomly pick a queue to run `Func`.
	Queues []string `json:"queues"`
	// Maximum number of concurrent instances for this job.
	// Minimum value is 1, cannot be set to 0 or negative.
	// Default: 1
	// Note: In protobuf, values ≤ 0 will be treated as 1.
	MaxInstances int `json:"max_instances"`

	// Automatic update, not manual setting.
	LastRunTime time.Time `json:"last_run_time"`
	// Automatic update, not manual setting.
	// When the job is paused, this field is set to `9999-09-09 09:09:09`.
	NextRunTime time.Time `json:"next_run_time"`
	// Optional: `JOB_STATUS_RUNNING` | `JOB_STATUS_PAUSED`
	// It should not be set manually.
	Status string `json:"status"`
}

// `sort.Interface`, sorted by 'NextRunTime', ascend.
type JobSlice []Job

func (js JobSlice) Len() int           { return len(js) }
func (js JobSlice) Less(i, j int) bool { return js[i].NextRunTime.Before(js[j].NextRunTime) }
func (js JobSlice) Swap(i, j int)      { js[i], js[j] = js[j], js[i] }

func (j *Job) setId() {
	j.Id = strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
}

// Initialization functions for each job,
// called when the scheduler run `AddJob`.
func (j *Job) init() error {
	j.setId()

	j.Status = JOB_STATUS_RUNNING
	j.setDefaults()

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

// setDefaults fills execution defaults without changing the job identity.
// It is shared by jobs being persisted and jobs submitted directly for execution.
func (j *Job) setDefaults() {
	if j.Timezone == "" {
		j.Timezone = "UTC"
	}

	if j.FuncName == "" && j.Func != nil {
		j.FuncName = getFuncName(j.Func)
	}

	if j.Args == nil {
		j.Args = map[string]any{}
	}

	if j.Timeout == "" {
		j.Timeout = "1h"
	}

	if j.Queues == nil {
		j.Queues = []string{}
	}

	if j.MaxInstances <= 0 {
		j.MaxInstances = 1
	}
}

// prepareForRun applies defaults and validates a job without generating a new ID
// or recalculating its schedule. Direct execution and queued execution both use
// the same validation so malformed remote requests cannot reach function lookup.
func (j *Job) prepareForRun() error {
	j.setDefaults()
	return j.check()
}

// Called when the job run `init` or scheduler run `UpdateJob`.
func (j *Job) check() error {
	if _, ok := getRegisteredFunc(j.FuncName); !ok {
		return FuncUnregisteredError(j.FuncName)
	}

	if _, err := j.timeoutDuration(); err != nil {
		return err
	}

	if j.MaxInstances <= 0 {
		return fmt.Errorf("job `%s` MaxInstances must be greater than 0, got %d", j.FullName(), j.MaxInstances)
	}

	return nil
}

func (j Job) timeoutDuration() (time.Duration, error) {
	timeout, err := time.ParseDuration(j.Timeout)
	if err != nil {
		return 0, &JobTimeoutError{FullName: j.FullName(), Timeout: j.Timeout, Err: err}
	}
	if timeout <= 0 {
		return 0, &JobTimeoutError{
			FullName: j.FullName(),
			Timeout:  j.Timeout,
			Err:      fmt.Errorf("must be greater than zero"),
		}
	}

	return timeout, nil
}

func (j Job) intervalDuration() (time.Duration, error) {
	interval, err := time.ParseDuration(j.Interval)
	if err != nil {
		return 0, fmt.Errorf("job `%s` Interval `%s` error: %s", j.FullName(), j.Interval, err)
	}
	if interval < time.Second || interval%time.Second != 0 {
		return 0, fmt.Errorf(
			"job `%s` Interval `%s` error: must be a positive whole number of seconds",
			j.FullName(), j.Interval,
		)
	}

	return interval, nil
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
			"'FuncName':'%s', 'Args':'%s', 'Timeout':'%s', 'Queues':'%s', 'MaxInstances':'%d', "+
			"'LastRunTime':'%s', 'NextRunTime':'%s', 'Status':'%s'}",
		j.Id, j.Name, j.Type, j.StartAt, j.EndAt,
		j.Interval, j.CronExpr, j.Timezone,
		j.FuncName, j.Args, j.Timeout, j.Queues, j.MaxInstances,
		j.LastRunTimeWithTimezone(), j.NextRunTimeWithTimezone(), j.Status,
	)
}

func (j Job) DeepCopy() (Job, error) {
	bJ, err := JobMarshal(j)
	if err != nil {
		return Job{}, err
	}
	cJ, err := JobUnmarshal(bJ)
	if err != nil {
		return Job{}, err
	}
	return cJ, nil
}

// Serialize Job and convert to Bytes
func JobMarshal(j Job) ([]byte, error) {
	return json.Marshal(j)
}

// Deserialize Bytes and convert to Job
func JobUnmarshal(bJ []byte) (Job, error) {
	var j Job
	err := json.Unmarshal(bJ, &j)
	if err != nil {
		return Job{}, err
	}
	return j, nil
}

type FuncPkg struct {
	Func func(context.Context, Job) (result string)
	// About this function.
	Info string
}

// Record the actual path of function and the corresponding function.
// Since golang can't serialize functions,
// need to register them with `RegisterFuncs` before using it.
var (
	FuncMap  = make(map[string]FuncPkg)
	funcMapM sync.RWMutex
)

func getRegisteredFunc(name string) (FuncPkg, bool) {
	funcMapM.RLock()
	defer funcMapM.RUnlock()

	fp, ok := FuncMap[name]
	return fp, ok
}

func FuncMapReadable() []map[string]string {
	funcs := []map[string]string{}

	funcMapM.RLock()
	defer funcMapM.RUnlock()

	for fName, fPkg := range FuncMap {
		funcs = append(funcs, map[string]string{
			"name": fName, "info": fPkg.Info,
		})
	}

	return funcs
}

func getFuncName(f func(context.Context, Job) (result string)) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func RegisterFuncs(fps ...FuncPkg) {
	funcMapM.Lock()
	defer funcMapM.Unlock()

	for _, fp := range fps {
		fName := getFuncName(fp.Func)
		FuncMap[fName] = fp
	}
}
