package agscheduler

import "fmt"

type JobNotFoundError string
type FuncUnregisteredError string

func (e JobNotFoundError) Error() string {
	return fmt.Sprintf("job `%s` not found!", string(e))
}

func (e FuncUnregisteredError) Error() string {
	return fmt.Sprintf("function `%s` unregistered!", string(e))
}
