// go run rpc.go

package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/services"
	pb "github.com/kwkwc/agscheduler/services/proto"
	"github.com/kwkwc/agscheduler/stores"
)

var ctx = context.Background()

func printMsg(j agscheduler.Job) {
	slog.Info(fmt.Sprintf("Run job `%s` %s\n", j.FullName(), j.Args))
}

func runExampleRPC(c pb.SchedulerClient) {
	agscheduler.RegisterFuncs(printMsg)

	job1 := agscheduler.Job{
		Name:     "Job1",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "2s",
		Timezone: "UTC",
		FuncName: "main.printMsg",
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	pbJob1, _ := c.AddJob(ctx, agscheduler.JobToPbJobPtr(job1))
	job1 = agscheduler.PbJobPtrToJob(pbJob1)
	slog.Info(fmt.Sprintf("Scheduler add job `%s` %s.\n\n", job1.FullName(), job1))

	job2 := agscheduler.Job{
		Name:     "Job2",
		Type:     agscheduler.TYPE_CRON,
		CronExpr: "*/1 * * * *",
		Timezone: "Asia/Shanghai",
		FuncName: "main.printMsg",
		Args:     map[string]any{"arg4": "4", "arg5": "5", "arg6": "6", "arg7": "7"},
	}
	pbJob2, _ := c.AddJob(ctx, agscheduler.JobToPbJobPtr(job2))
	job2 = agscheduler.PbJobPtrToJob(pbJob2)
	slog.Info(fmt.Sprintf("Scheduler add job `%s` %s.\n\n", job2.FullName(), job2))

	c.Start(ctx, &emptypb.Empty{})
	slog.Info("Scheduler start.\n\n")

	job3 := agscheduler.Job{
		Name:     "Job3",
		Type:     agscheduler.TYPE_DATETIME,
		StartAt:  "2023-09-22 07:30:08",
		Timezone: "America/New_York",
		FuncName: "main.printMsg",
		Args:     map[string]any{"arg8": "8", "arg9": "9"},
	}
	pbJob3, _ := c.AddJob(ctx, agscheduler.JobToPbJobPtr(job3))
	job3 = agscheduler.PbJobPtrToJob(pbJob3)
	slog.Info(fmt.Sprintf("Scheduler add job `%s` %s.\n\n", job3.FullName(), job3))

	pbJobs, _ := c.GetAllJobs(ctx, &emptypb.Empty{})
	jobs := agscheduler.PbJobsPtrToJobs(pbJobs)
	slog.Info(fmt.Sprintf("Scheduler get all jobs %s.\n\n", jobs))

	slog.Info("Sleep 10s......\n\n")
	time.Sleep(10 * time.Second)

	pbJob1, _ = c.GetJob(ctx, &pb.JobId{Id: job1.Id})
	job1 = agscheduler.PbJobPtrToJob(pbJob1)
	slog.Info(fmt.Sprintf("Scheduler get job `%s` %s.\n\n", job1.FullName(), job1))

	job2.Type = agscheduler.TYPE_INTERVAL
	job2.Interval = "4s"
	pbJob2, _ = c.UpdateJob(ctx, agscheduler.JobToPbJobPtr(job2))
	job2 = agscheduler.PbJobPtrToJob(pbJob2)
	slog.Info(fmt.Sprintf("Scheduler update job `%s` %s.\n\n", job2.FullName(), job2))

	slog.Info("Sleep 8s......")
	time.Sleep(8 * time.Second)

	pbJob1, _ = c.PauseJob(ctx, &pb.JobId{Id: job1.Id})
	job1 = agscheduler.PbJobPtrToJob(pbJob1)
	slog.Info(fmt.Sprintf("Scheduler pause job `%s`.\n\n", job1.FullName()))

	slog.Info("Sleep 6s......\n\n")
	time.Sleep(6 * time.Second)

	pbJob1, _ = c.ResumeJob(ctx, &pb.JobId{Id: job1.Id})
	job1 = agscheduler.PbJobPtrToJob(pbJob1)
	slog.Info(fmt.Sprintf("Scheduler resume job `%s`.\n\n", job1.FullName()))

	c.DeleteJob(ctx, &pb.JobId{Id: job2.Id})
	slog.Info(fmt.Sprintf("Scheduler delete job `%s`.\n\n", job2.FullName()))

	slog.Info("Sleep 6s......\n\n")
	time.Sleep(6 * time.Second)

	c.Stop(ctx, &emptypb.Empty{})
	slog.Info("Scheduler stop.\n\n")

	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	c.Start(ctx, &emptypb.Empty{})
	slog.Info("Scheduler start.\n\n")

	slog.Info("Sleep 4s......\n\n")
	time.Sleep(4 * time.Second)

	c.DeleteAllJobs(ctx, &emptypb.Empty{})
	slog.Info("Scheduler delete all jobs.\n\n")
}

func main() {
	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	service := services.SchedulerRPCService{Scheduler: scheduler}
	service.Start("")

	conn, _ := grpc.Dial("127.0.0.1:36363", grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := pb.NewSchedulerClient(conn)

	runExampleRPC(client)

	select {}
}