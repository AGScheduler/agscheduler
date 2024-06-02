package services

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/agscheduler/agscheduler"
	pb "github.com/agscheduler/agscheduler/services/proto"
)

type sGRPCService struct {
	pb.UnimplementedSchedulerServer

	scheduler *agscheduler.Scheduler
}

func (sgrs *sGRPCService) AddJob(ctx context.Context, pbJob *pb.Job) (*pb.Job, error) {
	j := agscheduler.PbJobPtrToJob(pbJob)
	j, err := sgrs.scheduler.AddJob(j)
	if err != nil {
		return &pb.Job{}, err
	}

	return agscheduler.JobToPbJobPtr(j)
}

func (sgrs *sGRPCService) GetJob(ctx context.Context, req *pb.JobReq) (*pb.Job, error) {
	j, err := sgrs.scheduler.GetJob(req.GetId())
	if err != nil {
		return &pb.Job{}, err
	}

	return agscheduler.JobToPbJobPtr(j)
}

func (sgrs *sGRPCService) GetAllJobs(ctx context.Context, in *emptypb.Empty) (*pb.JobsResp, error) {
	js, err := sgrs.scheduler.GetAllJobs()
	if err != nil {
		return &pb.JobsResp{}, err
	}

	pbJs, err := agscheduler.JobsToPbJobsPtr(js)
	if err != nil {
		return &pb.JobsResp{}, err
	}

	return &pb.JobsResp{Jobs: pbJs}, err
}

func (sgrs *sGRPCService) UpdateJob(ctx context.Context, pbJob *pb.Job) (*pb.Job, error) {
	j := agscheduler.PbJobPtrToJob(pbJob)
	j, err := sgrs.scheduler.UpdateJob(j)
	if err != nil {
		return &pb.Job{}, err
	}

	return agscheduler.JobToPbJobPtr(j)
}

func (sgrs *sGRPCService) DeleteJob(ctx context.Context, req *pb.JobReq) (*emptypb.Empty, error) {
	err := sgrs.scheduler.DeleteJob(req.GetId())
	return &emptypb.Empty{}, err
}

func (sgrs *sGRPCService) DeleteAllJobs(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	err := sgrs.scheduler.DeleteAllJobs()
	return &emptypb.Empty{}, err
}

func (sgrs *sGRPCService) PauseJob(ctx context.Context, req *pb.JobReq) (*pb.Job, error) {
	j, err := sgrs.scheduler.PauseJob(req.GetId())
	if err != nil {
		return &pb.Job{}, err
	}

	return agscheduler.JobToPbJobPtr(j)
}

func (sgrs *sGRPCService) ResumeJob(ctx context.Context, req *pb.JobReq) (*pb.Job, error) {
	j, err := sgrs.scheduler.ResumeJob(req.GetId())
	if err != nil {
		return &pb.Job{}, err
	}

	return agscheduler.JobToPbJobPtr(j)
}

func (sgrs *sGRPCService) RunJob(ctx context.Context, pbJob *pb.Job) (*emptypb.Empty, error) {
	j := agscheduler.PbJobPtrToJob(pbJob)
	err := sgrs.scheduler.RunJob(j)
	return &emptypb.Empty{}, err
}

func (sgrs *sGRPCService) ScheduleJob(ctx context.Context, pbJob *pb.Job) (*emptypb.Empty, error) {
	j := agscheduler.PbJobPtrToJob(pbJob)
	err := sgrs.scheduler.ScheduleJob(j)
	return &emptypb.Empty{}, err
}

func (sgrs *sGRPCService) Start(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	sgrs.scheduler.Start()
	return &emptypb.Empty{}, nil
}

func (sgrs *sGRPCService) Stop(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	sgrs.scheduler.Stop()
	return &emptypb.Empty{}, nil
}
