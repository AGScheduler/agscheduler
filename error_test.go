package agscheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobNotFoundError(t *testing.T) {
	err := JobNotFoundError("1")

	assert.Equal(t, "Job with id 1 not found!", err.Error())
}
