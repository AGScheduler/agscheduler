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

func (srs *sRPCService) ScheduleJob(ctx context.Context, pbJob *pb.Job) (*emptypb.Empty, error) {
	j := agscheduler.PbJobPtrToJob(pbJob)
	err := srs.scheduler.ScheduleJob(j)
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
			slog.Error(fmt.Sprintf("Scheduler gRPC Service method: `%s`, err: `%s`", info.FullMethod, err))
		}
	}()

	resp, err = handler(ctx, req)
	return resp, err
}

type SchedulerRPCService struct {
	Scheduler *agscheduler.Scheduler

	// Default: `127.0.0.1:36360`
	Address string

	srv *grpc.Server
}

func (s *SchedulerRPCService) Start() error {
	if s.Address == "" {
		s.Address = "127.0.0.1:36360"
	}

	lis, err := net.Listen("tcp", s.Address)
	if err != nil {
		return fmt.Errorf("scheduler gRPC Service listen failure: %s", err)
	}

	chap := &ClusterProxy{Scheduler: s.Scheduler}
	s.srv = grpc.NewServer(grpc.ChainUnaryInterceptor(panicInterceptor, chap.GRPCProxyInterceptor))
	pb.RegisterSchedulerServer(s.srv, &sRPCService{scheduler: s.Scheduler})
	slog.Info(fmt.Sprintf("Scheduler gRPC Service listening at: %s", lis.Addr()))

	go func() {
		if err := s.srv.Serve(lis); err != nil {
			slog.Error(fmt.Sprintf("Scheduler gRPC Service Unavailable: %s", err))
		}
	}()

	return nil
}

func (s *SchedulerRPCService) Stop() error {
	s.srv.Stop()

	return nil
}
