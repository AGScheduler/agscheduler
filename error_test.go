package agscheduler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobNotFoundError(t *testing.T) {
	err := JobNotFoundError("1")

	assert.Equal(t, "jobId `1` not found!", err.Error())
}

func TestFuncUnregisteredError(t *testing.T) {
	err := FuncUnregisteredError("func")

	assert.Equal(t, "function `func` unregistered!", err.Error())
}

func TestJobTimeoutError(t *testing.T) {
	err := &JobTimeoutError{FullName: "1:job", Timeout: "1s", Err: errors.New("err")}

	assert.Equal(t, "job `1:job` Timeout `1s` error: err!", err.Error())
}
