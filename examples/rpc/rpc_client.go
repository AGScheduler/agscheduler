// 1. go run examples/rpc/rpc_server.go
// 2. go run examples/rpc/rpc_client.go

package main

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kwkwc/agscheduler"
	pb "github.com/kwkwc/agscheduler/services/proto"
)

var ctx = context.Background()

func runExampleRPC(c pb.SchedulerClient) {
	job1 := agscheduler.Job{
		Name:     "Job1",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "2s",
		Timezone: "UTC",
		FuncName: "github.com/kwkwc/agscheduler/examples.PrintMsg",
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	pbJob1, _ := c.AddJob(ctx, agscheduler.JobToPbJobPtr(job1))
	job1 = agscheduler.PbJobPtrToJob(pbJob1)
	slog.Info(fmt.Sprintf("%s.\n\n", job1))
}

func main() {
	conn, _ := grpc.Dial("127.0.0.1:36363", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	client := pb.NewSchedulerClient(conn)

	runExampleRPC(client)
}
