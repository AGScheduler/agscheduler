package services

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/agscheduler/agscheduler"
	pb "github.com/agscheduler/agscheduler/services/proto"
)

type bGRPCService struct {
	pb.UnimplementedBaseServer

	scheduler *agscheduler.Scheduler
}

func (bgrs *bGRPCService) GetInfo(ctx context.Context, in *emptypb.Empty) (*pb.InfoResp, error) {
	info, err := structpb.NewStruct(bgrs.scheduler.Info())
	if err != nil {
		return nil, err
	}

	return &pb.InfoResp{Info: info}, nil
}

func (bgrs *bGRPCService) GetFuncs(ctx context.Context, in *emptypb.Empty) (*pb.FuncsResp, error) {
	pbFs := []*pb.Func{}

	fs := agscheduler.FuncMapReadable()
	for _, f := range fs {
		pbF := &pb.Func{Name: f["name"], Info: f["Info"]}
		pbFs = append(pbFs, pbF)
	}

	return &pb.FuncsResp{Funcs: pbFs}, nil
}

type GRPCService struct {
	Scheduler *agscheduler.Scheduler

	// Default: `127.0.0.1:36360`
	Address string
	// SHA256 encrypted authorization password, e.g. here is admin
	// echo -n admin | shasum -a 256
	// `8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918`
	//
	// Add `auth-password-sha2` metadata on request.
	PasswordSha2 string

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
	s.srv = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.panicInterceptor, s.verifyPassword, cp.gRPCProxyInterceptor,
		),
	)

	bgrs := &bGRPCService{scheduler: s.Scheduler}
	pb.RegisterBaseServer(s.srv, bgrs)

	sgrs := &sGRPCService{scheduler: s.Scheduler}
	pb.RegisterSchedulerServer(s.srv, sgrs)

	if s.Scheduler.IsClusterMode() {
		cgrs := &cGRPCService{cn: agscheduler.GetClusterNode(s.Scheduler)}
		pb.RegisterClusterServer(s.srv, cgrs)
	}

	if s.Scheduler.HasRecorder() {
		rgrs := &rGRPCService{recorder: agscheduler.GetRecorder(s.Scheduler)}
		pb.RegisterRecorderServer(s.srv, rgrs)
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

func (s *GRPCService) panicInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("gRPC Service method: `%s`, err: `%s`", info.FullMethod, err))
		}
	}()

	return handler(ctx, req)
}

func (s *GRPCService) verifyPassword(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	err = gRPCVerifyPassword(ctx, s.PasswordSha2)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}
