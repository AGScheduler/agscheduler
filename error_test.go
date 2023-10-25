package agscheduler

import (
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
