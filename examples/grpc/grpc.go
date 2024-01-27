// go run examples/grpc/grpc.go

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/examples"
	"github.com/kwkwc/agscheduler/services"
	pb "github.com/kwkwc/agscheduler/services/proto"
	"github.com/kwkwc/agscheduler/stores"
)

func runExampleGRPC(c pb.SchedulerClient) {
	ctx := context.Background()

	job1 := agscheduler.Job{
		Name:     "Job1",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "2s",
		Timezone: "UTC",
		FuncName: "github.com/kwkwc/agscheduler/examples.PrintMsg",
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	pbJob1, _ := agscheduler.JobToPbJobPtr(job1)
	pbJob1, _ = c.AddJob(ctx, pbJob1)
	job1 = agscheduler.PbJobPtrToJob(pbJob1)
	slog.Info(fmt.Sprintf("%s.\n\n", job1))

	job2 := agscheduler.Job{
		Name:     "Job2",
		Type:     agscheduler.TYPE_CRON,
		CronExpr: "*/1 * * * *",
		Timezone: "Asia/Shanghai",
		FuncName: "github.com/kwkwc/agscheduler/examples.PrintMsg",
		Args:     map[string]any{"arg4": "4", "arg5": "5", "arg6": "6", "arg7": "7"},
	}
	pbJob2, _ := agscheduler.JobToPbJobPtr(job2)
	pbJob2, _ = c.AddJob(ctx, pbJob2)
	job2 = agscheduler.PbJobPtrToJob(pbJob2)
	slog.Info(fmt.Sprintf("%s.\n\n", job2))

	c.Start(ctx, &emptypb.Empty{})

	job3 := agscheduler.Job{
		Name:     "Job3",
		Type:     agscheduler.TYPE_DATETIME,
		StartAt:  "2023-09-22 07:30:08",
		Timezone: "America/New_York",
		FuncName: "github.com/kwkwc/agscheduler/examples.PrintMsg",
		Args:     map[string]any{"arg8": "8", "arg9": "9"},
	}
	pbJob3, _ := agscheduler.JobToPbJobPtr(job3)
	pbJob3, _ = c.AddJob(ctx, pbJob3)
	job3 = agscheduler.PbJobPtrToJob(pbJob3)
	slog.Info(fmt.Sprintf("%s.\n\n", job3))

	pbJobs, _ := c.GetAllJobs(ctx, &emptypb.Empty{})
	jobs := agscheduler.PbJobsPtrToJobs(pbJobs)
	slog.Info(fmt.Sprintf("Scheduler get all jobs %s.\n\n", jobs))

	slog.Info("Sleep 5s......\n\n")
	time.Sleep(5 * time.Second)

	pbJob1, _ = c.GetJob(ctx, &pb.JobId{Id: job1.Id})
	job1 = agscheduler.PbJobPtrToJob(pbJob1)
	slog.Info(fmt.Sprintf("Scheduler get job `%s` %s.\n\n", job1.FullName(), job1))

	job2.Type = agscheduler.TYPE_INTERVAL
	job2.Interval = "3s"
	pbJob2, _ = agscheduler.JobToPbJobPtr(job2)
	pbJob2, _ = c.UpdateJob(ctx, pbJob2)
	job2 = agscheduler.PbJobPtrToJob(pbJob2)
	slog.Info(fmt.Sprintf("Scheduler update job `%s` %s.\n\n", job2.FullName(), job2))

	slog.Info("Sleep 4s......")
	time.Sleep(4 * time.Second)

	pbJob1, _ = c.PauseJob(ctx, &pb.JobId{Id: job1.Id})
	job1 = agscheduler.PbJobPtrToJob(pbJob1)

	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	pbJob1, _ = c.ResumeJob(ctx, &pb.JobId{Id: job1.Id})
	job1 = agscheduler.PbJobPtrToJob(pbJob1)

	c.DeleteJob(ctx, &pb.JobId{Id: job2.Id})

	slog.Info("Sleep 4s......\n\n")
	time.Sleep(4 * time.Second)

	c.Stop(ctx, &emptypb.Empty{})

	c.RunJob(ctx, pbJob1)

	c.ScheduleJob(ctx, pbJob1)

	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	c.Start(ctx, &emptypb.Empty{})

	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	c.DeleteAllJobs(ctx, &emptypb.Empty{})

	c.Stop(ctx, &emptypb.Empty{})
}

func main() {
	agscheduler.RegisterFuncs(examples.PrintMsg)

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	grservice := services.GRPCService{
		Scheduler: scheduler,
		Address:   "127.0.0.1:36360",
	}
	err = grservice.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start service: %s", err))
		os.Exit(1)
	}

	conn, _ := grpc.Dial("127.0.0.1:36360", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	client := pb.NewSchedulerClient(conn)

	runExampleGRPC(client)

	err = grservice.Stop()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to stop service: %s", err))
		os.Exit(1)
	}
}
