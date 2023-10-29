package examples

import (
	"fmt"
	"log/slog"

	"github.com/kwkwc/agscheduler"
)

func PrintMsg(j agscheduler.Job) {
	slog.Info(fmt.Sprintf("Run job `%s` %s\n\n", j.FullName(), j.Args))
}
