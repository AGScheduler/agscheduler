package agscheduler

import "fmt"

type JobNotFoundError string

func (j JobNotFoundError) Error() string {
	return fmt.Sprintf("Job with id %s not found!", string(j))
}
