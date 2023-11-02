package services

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/kwkwc/agscheduler"
	pb "github.com/kwkwc/agscheduler/services/proto"
)

type sRPCService struct {
	pb.UnimplementedSchedulerServer

	scheduler *agscheduler.Scheduler
}

func (srs *sRPCService) AddJob(ctx context.Context, pbJob *pb.Job) (*pb.Job, error) {
	j := agscheduler.PbJobPtrToJob(pbJob)
	j, err := srs.scheduler.AddJob(j)
	return agscheduler.JobToPbJobPtr(j), err
}

func (srs *sRPCService) GetJob(ctx context.Context, jobId *pb.JobId) (*pb.Job, error) {
	j, err := srs.scheduler.GetJob(jobId.GetId())
	return agscheduler.JobToPbJobPtr(j), err
}

func (srs *sRPCService) GetAllJobs(ctx context.Context, in *emptypb.Empty) (*pb.Jobs, error) {
	js, err := srs.scheduler.GetAllJobs()
	return agscheduler.JobsToPbJobsPtr(js), err
}

func (srs *sRPCService) UpdateJob(ctx context.Context, pbJob *pb.Job) (*pb.Job, error) {
	j := agscheduler.PbJobPtrToJob(pbJob)
	j, err := srs.scheduler.UpdateJob(j)
	return agscheduler.JobToPbJobPtr(j), err
}

func (srs *sRPCService) DeleteJob(ctx context.Context, jobId *pb.JobId) (*emptypb.Empty, error) {
	err := srs.scheduler.DeleteJob(jobId.GetId())
	return &emptypb.Empty{}, err
}

func (srs *sRPCService) DeleteAllJobs(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	err := srs.scheduler.DeleteAllJobs()
	return &emptypb.Empty{}, err
}

func (srs *sRPCService) PauseJob(ctx context.Context, jobId *pb.JobId) (*pb.Job, error) {
	j, err := srs.scheduler.PauseJob(jobId.GetId())
	return agscheduler.JobToPbJobPtr(j), err
}

func (srs *sRPCService) ResumeJob(ctx context.Context, jobId *pb.JobId) (*pb.Job, error) {
	j, err := srs.scheduler.ResumeJob(jobId.GetId())
	return agscheduler.JobToPbJobPtr(j), err
}

func (srs *sRPCService) RunJob(ctx context.Context, pbJob *pb.Job) (*emptypb.Empty, error) {
	j := agscheduler.PbJobPtrToJob(pbJob)
	err := srs.scheduler.RunJob(j)
	return &emptypb.Empty{}, err
}

func (srs *sRPCService) Start(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	srs.scheduler.Start()
	return &emptypb.Empty{}, nil
}

func (srs *sRPCService) Stop(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	srs.scheduler.Stop()
	return &emptypb.Empty{}, nil
}

func panicInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("Scheduler RPC Service method: `%s`, err: `%s`", info.FullMethod, err))
		}
	}()

	resp, err = handler(ctx, req)
	return resp, err
}

type SchedulerRPCService struct {
	Scheduler *agscheduler.Scheduler
	Address   string
}

func (s *SchedulerRPCService) Start() error {
	if s.Address == "" {
		s.Address = "127.0.0.1:36363"
	}

	lis, err := net.Listen("tcp", s.Address)
	if err != nil {
		return fmt.Errorf("scheduler RPC Service listen failure: %s", err)
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(panicInterceptor))
	pb.RegisterSchedulerServer(srv, &sRPCService{scheduler: s.Scheduler})
	slog.Info(fmt.Sprintf("Scheduler RPC Service listening at: %s", lis.Addr()))

	go func() {
		if err := srv.Serve(lis); err != nil {
			slog.Error(fmt.Sprintf("Scheduler RPC Service Unavailable: %s", err))
		}
	}()

	return nil
}
