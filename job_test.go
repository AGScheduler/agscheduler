package agscheduler

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/kwkwc/agscheduler/services/proto"
)

func getJob() Job {
	return Job{
		Name:     "Job",
		Type:     TYPE_INTERVAL,
		Interval: "1s",
		Func:     func(j Job) {},
		Args:     map[string]any{},
	}
}

func TestJobSetId(t *testing.T) {
	j := getJob()
	j.SetId()

	assert.Len(t, j.Id, 16)
}

func TestJobString(t *testing.T) {
	j := getJob()
	typeOfJob := reflect.TypeOf(j)
	for i := 0; i < typeOfJob.NumField(); i++ {
		fieldType := typeOfJob.Field(i)
		if fieldType.Name == "Func" {
			continue
		}
		assert.Contains(t, j.String(), "'"+fieldType.Name+"'")
	}
}

func TestJobStateDump(t *testing.T) {
	j := getJob()
	state, err := StateDump(j)

	assert.IsType(t, []byte{}, state)
	assert.NotEmpty(t, state)
	assert.NoError(t, err)
}

func TestJobStateLoad(t *testing.T) {
	j := getJob()
	state, _ := StateDump(j)
	j, err := StateLoad(state)

	assert.IsType(t, Job{}, j)
	assert.NotEmpty(t, j)
	assert.NoError(t, err)
}

func TestJobStateLoadError(t *testing.T) {
	j, err := StateLoad([]byte("job"))

	assert.Empty(t, j)
	assert.Error(t, err)
}

func TestJobToPbJobPtr(t *testing.T) {
	j := getJob()
	pbJ := JobToPbJobPtr(j)

	assert.IsType(t, &pb.Job{}, pbJ)
	assert.NotEmpty(t, pbJ)
}

func TestPbJobPtrToJob(t *testing.T) {
	j := getJob()
	pbJ := JobToPbJobPtr(j)
	j = PbJobPtrToJob(pbJ)

	assert.IsType(t, Job{}, j)
	assert.NotEmpty(t, j)
}

func TestJobsToPbJobsPtr(t *testing.T) {
	js := make([]Job, 0)
	js = append(js, getJob())
	js = append(js, getJob())
	pbJs := JobsToPbJobsPtr(js)

	assert.IsType(t, &pb.Jobs{}, pbJs)
	assert.Len(t, pbJs.Jobs, 2)
}

func TestPbJobsPtrToJobs(t *testing.T) {
	js := make([]Job, 0)
	js = append(js, getJob())
	js = append(js, getJob())
	pbJs := JobsToPbJobsPtr(js)
	js = PbJobsPtrToJobs(pbJs)

	assert.IsType(t, []Job{}, js)
	assert.Len(t, js, 2)
}

func TestRegisterFuncs(t *testing.T) {
	assert.Empty(t, funcMap)

	RegisterFuncs(func(j Job) {})

	assert.Len(t, funcMap, 1)
}
