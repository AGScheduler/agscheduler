package agscheduler

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestRegisterFuncs(t *testing.T) {
	fp := FuncPkg{
		Func: func(ctx context.Context, j Job) (result string) { return },
		Info: "test function",
	}
	RegisterFuncs(fp)

	registered, ok := getRegisteredFunc(getFuncName(fp.Func))
	assert.True(t, ok)
	assert.Equal(t, fp.Info, registered.Info)
}

func TestFuncMapReadable(t *testing.T) {
	fp := FuncPkg{
		Func: func(ctx context.Context, j Job) (result string) { return },
		Info: "readable function",
	}
	RegisterFuncs(fp)

	assert.Contains(t, FuncMapReadable(), map[string]string{
		"name": getFuncName(fp.Func),
		"info": fp.Info,
	})
}

func TestFuncMapConcurrentAccess(t *testing.T) {
	fp := FuncPkg{
		Func: func(ctx context.Context, j Job) (result string) { return },
		Info: "concurrent function",
	}
	RegisterFuncs(fp)
	name := getFuncName(fp.Func)

	var wg sync.WaitGroup
	wg.Add(4)
	for range 3 {
		go func() {
			defer wg.Done()
			for range 1000 {
				if _, ok := getRegisteredFunc(name); !ok {
					t.Error("registered function not found")
					return
				}
				_ = FuncMapReadable()
			}
		}()
	}
	go func() {
		defer wg.Done()
		for range 1000 {
			RegisterFuncs(fp)
		}
	}()
	wg.Wait()
}
