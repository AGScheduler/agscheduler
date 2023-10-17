package agscheduler

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
	job := getJob()
	job.SetId()

	assert.Len(t, job.Id, 32)
}

func TestJobString(t *testing.T) {
	job := getJob()
	typeOfJob := reflect.TypeOf(job)
	for i := 0; i < typeOfJob.NumField(); i++ {
		fieldType := typeOfJob.Field(i)
		if fieldType.Name == "Func" {
			continue
		}
		assert.Contains(t, job.String(), "'"+fieldType.Name+"'")
	}
}

func TestJobStateDumps(t *testing.T) {
	job := getJob()
	state, err := StateDumps(job)

	assert.IsType(t, []byte{}, state)
	assert.NotEmpty(t, state)
	assert.NoError(t, err)
}

func TestJobStateLoads(t *testing.T) {
	job := getJob()
	state, _ := StateDumps(job)
	job, err := StateLoads(state)

	assert.IsType(t, Job{}, job)
	assert.NotEmpty(t, job)
	assert.NoError(t, err)
}

func TestJobStateLoadsError(t *testing.T) {
	job, err := StateLoads([]byte("job"))

	assert.Empty(t, job)
	assert.Error(t, err)
}

func TestRegisterFuncs(t *testing.T) {
	assert.Empty(t, funcMap)

	RegisterFuncs(func(j Job) {})

	assert.Len(t, funcMap, 1)
}
