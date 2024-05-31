package examples

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/agscheduler/agscheduler"
)

func PrintMsg(ctx context.Context, j agscheduler.Job) (result []byte) {
	slog.Info(fmt.Sprintf("Run job `%s` %s\n\n", j.FullName(), j.Args))
	return
}

func PrintMsgSleep(ctx context.Context, j agscheduler.Job) (result []byte) {
	slog.Info(fmt.Sprintf("Run job `%s`, sleep 2s\n\n", j.FullName()))
	time.Sleep(2 * time.Second)
	return
}
