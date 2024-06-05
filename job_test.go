package agscheduler

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/agscheduler/agscheduler/services/proto"
)

func getJob() Job {
	return Job{
		Name:     "Job",
		Type:     JOB_TYPE_INTERVAL,
		Interval: "1s",
		Func:     func(ctx context.Context, j Job) (result string) { return },
		Args:     map[string]any{},
	}
}

func TestJobSetId(t *testing.T) {
	j := getJob()
	j.setId()

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

func TestJobDeepCopy(t *testing.T) {
	j := getJob()
	cJ, err := j.DeepCopy()
	assert.NoError(t, err)

	cJ.Args["name"] = "test"
	assert.NotEmpty(t, cJ.Args)
	assert.Empty(t, j.Args)
}

func TestJobMarshal(t *testing.T) {
	j := getJob()
	bJ, err := JobMarshal(j)

	assert.IsType(t, []byte{}, bJ)
	assert.NotEmpty(t, bJ)
	assert.NoError(t, err)
}

func TestJobUnmarshal(t *testing.T) {
	j := getJob()
	bJ, err := JobMarshal(j)
	assert.NoError(t, err)
	j, err = JobUnmarshal(bJ)
	assert.NoError(t, err)

	assert.IsType(t, Job{}, j)
	assert.NotEmpty(t, j)
}

func TestJobUnmarshalError(t *testing.T) {
	j, err := JobUnmarshal([]byte("job"))
	assert.Error(t, err)

	assert.Empty(t, j)
}

func TestJobToPbJobPtr(t *testing.T) {
	j := getJob()
	pbJ, err := JobToPbJobPtr(j)
	assert.NoError(t, err)

	assert.IsType(t, &pb.Job{}, pbJ)
	assert.NotEmpty(t, pbJ)
}

func TestPbJobPtrToJob(t *testing.T) {
	j := getJob()
	pbJ, err := JobToPbJobPtr(j)
	assert.NoError(t, err)
	j = PbJobPtrToJob(pbJ)

	assert.IsType(t, Job{}, j)
	assert.NotEmpty(t, j)
}

func TestJobsToPbJobsPtr(t *testing.T) {
	js := make([]Job, 0)
	js = append(js, getJob())
	js = append(js, getJob())
	pbJs, err := JobsToPbJobsPtr(js)
	assert.NoError(t, err)

	assert.IsType(t, []*pb.Job{}, pbJs)
	assert.Len(t, pbJs, 2)
}

func TestPbJobsPtrToJobs(t *testing.T) {
	js := make([]Job, 0)
	js = append(js, getJob())
	js = append(js, getJob())
	pbJs, err := JobsToPbJobsPtr(js)
	assert.NoError(t, err)
	js = PbJobsPtrToJobs(pbJs)

	assert.IsType(t, []Job{}, js)
	assert.Len(t, js, 2)
}

func TestRegisterFuncs(t *testing.T) {
	assert.Empty(t, FuncMap)

	RegisterFuncs(
		FuncPkg{Func: func(ctx context.Context, j Job) (result string) { return }},
	)

	assert.Len(t, FuncMap, 1)
}

func TestFuncMapReadable(t *testing.T) {
	RegisterFuncs(
		FuncPkg{Func: func(ctx context.Context, j Job) (result string) { return }},
	)
	funcLen := len(FuncMap)

	assert.Len(t, FuncMapReadable(), funcLen)
}
