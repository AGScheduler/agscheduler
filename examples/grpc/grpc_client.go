// 1. go run examples/grpc/grpc_server.go
// 2. go run examples/grpc/grpc_client.go

package main

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/agscheduler/agscheduler"
	pb "github.com/agscheduler/agscheduler/services/proto"
)

func runExampleGRPC(c pb.SchedulerClient) {
	ctx := context.Background()
	// ctx = metadata.AppendToOutgoingContext(ctx, "auth-password-sha2", "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918")

	job1 := agscheduler.Job{
		Name:     "Job1",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "2s",
		Timezone: "UTC",
		FuncName: "github.com/agscheduler/agscheduler/examples.PrintMsg",
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	pbJob1, _ := agscheduler.JobToPbJobPtr(job1)
	pbJob1, _ = c.AddJob(ctx, pbJob1)
	job1 = agscheduler.PbJobPtrToJob(pbJob1)
	slog.Info(fmt.Sprintf("%s.\n\n", job1))

	c.Start(ctx, &emptypb.Empty{})
}

func main() {
	conn, _ := grpc.NewClient("127.0.0.1:36360", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	client := pb.NewSchedulerClient(conn)

	runExampleGRPC(client)
}
