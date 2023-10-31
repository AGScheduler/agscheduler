package agscheduler

import "fmt"

type JobNotFoundError string
type FuncUnregisteredError string

type JobTimeoutError struct {
	FullName string
	Timeout  string
	Err      error
}

func (e JobNotFoundError) Error() string {
	return fmt.Sprintf("jobId `%s` not found!", string(e))
}

func (e FuncUnregisteredError) Error() string {
	return fmt.Sprintf("function `%s` unregistered!", string(e))
}

func (e *JobTimeoutError) Error() string {
	return fmt.Sprintf("job `%s` Timeout `%s` error: %s!", e.FullName, e.Timeout, e.Err)
}
