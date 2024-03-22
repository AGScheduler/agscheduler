package services

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/kwkwc/agscheduler"
	pb "github.com/kwkwc/agscheduler/services/proto"
)

type bGRPCService struct {
	pb.UnimplementedBaseServer

	scheduler *agscheduler.Scheduler
}

func (bgrs *bGRPCService) GetInfo(ctx context.Context, in *emptypb.Empty) (*pb.Info, error) {
	info, err := structpb.NewStruct(bgrs.scheduler.Info())
	if err != nil {
		return nil, err
	}

	return &pb.Info{Info: info}, nil
}

func (bgrs *bGRPCService) GetFuncs(ctx context.Context, in *emptypb.Empty) (*pb.Funcs, error) {
	pbFs := &pb.Funcs{}

	fs := agscheduler.FuncMapReadable()
	for _, f := range fs {
		pbF := &pb.Func{Name: f["name"], Info: f["Info"]}
		pbFs.Funcs = append(pbFs.Funcs, pbF)
	}

	return pbFs, nil
}

func panicInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("gRPC Service method: `%s`, err: `%s`", info.FullMethod, err))
		}
	}()

	resp, err = handler(ctx, req)
	return resp, err
}

type GRPCService struct {
	Scheduler *agscheduler.Scheduler

	// Default: `127.0.0.1:36360`
	Address string

	srv *grpc.Server
}

func (s *GRPCService) Start() error {
	if s.Address == "" {
		s.Address = "127.0.0.1:36360"
	}

	lis, err := net.Listen("tcp", s.Address)
	if err != nil {
		return fmt.Errorf("gRPC Service listen failure: %s", err)
	}

	cp := &ClusterProxy{Scheduler: s.Scheduler}
	s.srv = grpc.NewServer(grpc.ChainUnaryInterceptor(panicInterceptor, cp.GRPCProxyInterceptor))

	bgrs := &bGRPCService{scheduler: s.Scheduler}
	pb.RegisterBaseServer(s.srv, bgrs)

	sgrs := &sGRPCService{scheduler: s.Scheduler}
	pb.RegisterSchedulerServer(s.srv, sgrs)

	if s.Scheduler.IsClusterMode() {
		cgrs := &cGRPCService{cn: agscheduler.GetClusterNode(s.Scheduler)}
		pb.RegisterClusterServer(s.srv, cgrs)
	}

	slog.Info(fmt.Sprintf("gRPC Service listening at: %s", lis.Addr()))

	go func() {
		if err := s.srv.Serve(lis); err != nil {
			slog.Error(fmt.Sprintf("gRPC Service Unavailable: %s", err))
		}
	}()

	return nil
}

func (s *GRPCService) Stop() error {
	slog.Info("gRPC Service stop")

	s.srv.Stop()

	return nil
}
