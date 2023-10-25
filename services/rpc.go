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

type RPCService struct {
	pb.UnimplementedSchedulerServer

	scheduler *agscheduler.Scheduler
}

func (rs *RPCService) AddJob(ctx context.Context, pbJob *pb.Job) (*pb.Job, error) {
	j := agscheduler.PbJobPtrToJob(pbJob)
	j, err := rs.scheduler.AddJob(j)
	return agscheduler.JobToPbJobPtr(j), err
}

func (rs *RPCService) GetJob(ctx context.Context, jobId *pb.JobId) (*pb.Job, error) {
	j, err := rs.scheduler.GetJob(jobId.GetId())
	return agscheduler.JobToPbJobPtr(j), err
}

func (rs *RPCService) GetAllJobs(ctx context.Context, in *emptypb.Empty) (*pb.Jobs, error) {
	js, err := rs.scheduler.GetAllJobs()
	return agscheduler.JobsToPbJobsPtr(js), err
}

func (rs *RPCService) UpdateJob(ctx context.Context, pbJob *pb.Job) (*pb.Job, error) {
	j := agscheduler.PbJobPtrToJob(pbJob)
	j, err := rs.scheduler.UpdateJob(j)
	return agscheduler.JobToPbJobPtr(j), err
}

func (rs *RPCService) DeleteJob(ctx context.Context, jobId *pb.JobId) (*emptypb.Empty, error) {
	err := rs.scheduler.DeleteJob(jobId.GetId())
	return &emptypb.Empty{}, err
}

func (rs *RPCService) DeleteAllJobs(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	err := rs.scheduler.DeleteAllJobs()
	return &emptypb.Empty{}, err
}

func (rs *RPCService) PauseJob(ctx context.Context, jobId *pb.JobId) (*pb.Job, error) {
	j, err := rs.scheduler.PauseJob(jobId.GetId())
	return agscheduler.JobToPbJobPtr(j), err
}

func (rs *RPCService) ResumeJob(ctx context.Context, jobId *pb.JobId) (*pb.Job, error) {
	j, err := rs.scheduler.ResumeJob(jobId.GetId())
	return agscheduler.JobToPbJobPtr(j), err
}

func (rs *RPCService) RunJob(ctx context.Context, jobId *pb.JobId) (*emptypb.Empty, error) {
	err := rs.scheduler.RunJob(jobId.GetId())
	return &emptypb.Empty{}, err
}

func (rs *RPCService) Start(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	rs.scheduler.Start()
	return &emptypb.Empty{}, nil
}

func (rs *RPCService) Stop(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	rs.scheduler.Stop()
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
}

func (s *SchedulerRPCService) Start(address string) error {
	if address == "" {
		address = "127.0.0.1:36363"
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("scheduler RPC Service listen failure: %s", err)
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(panicInterceptor))
	pb.RegisterSchedulerServer(srv, &RPCService{scheduler: s.Scheduler})
	slog.Info(fmt.Sprintf("Scheduler RPC Service listening at: %s", lis.Addr()))

	go func() {
		if err := srv.Serve(lis); err != nil {
			slog.Error(fmt.Sprintf("Scheduler RPC Service Unavailable: %s", err))
		}
	}()

	return nil
}
