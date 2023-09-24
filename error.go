package agscheduler

import "fmt"

type JobNotFound string

func (j JobNotFound) Error() string {
	return fmt.Sprintf("Job with id %s not found!", string(j))
}
