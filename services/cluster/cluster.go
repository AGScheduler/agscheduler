package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/agscheduler/agscheduler"
	grpcservice "github.com/agscheduler/agscheduler/services/grpc"
	httpservice "github.com/agscheduler/agscheduler/services/http"
)

type ClusterService struct {
	Cn           *agscheduler.ClusterNode
	PasswordSha2 string

	grs *grpcservice.GRPCService
	hs  *httpservice.HTTPService
	crs *clusterRPCService
}

func (s *ClusterService) Start() error {
	s.grs = &grpcservice.GRPCService{
		Scheduler:    s.Cn.Scheduler,
		PasswordSha2: s.PasswordSha2,
	}
	s.grs.Address = s.Cn.EndpointGRPC
	err := s.grs.Start()
	if err != nil {
		return err
	}

	s.hs = &httpservice.HTTPService{
		Scheduler:    s.Cn.Scheduler,
		PasswordSha2: s.PasswordSha2,
	}
	s.hs.Address = s.Cn.EndpointHTTP
	err = s.hs.Start()
	if err != nil {
		return err
	}

	s.crs = &clusterRPCService{Cn: s.Cn}
	err = s.crs.Start()
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Cluster Node Queue: `%s`", s.Cn.Queue))

	if !s.Cn.IsMainNode() {
		err = s.Cn.RegisterNodeRemote(context.TODO())
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to register node remote: %s", err))
		}
	}

	return nil
}

func (s *ClusterService) Stop() error {
	if s.grs != nil {
		if err := s.grs.Stop(); err != nil {
			return err
		}
		s.grs = nil
	}

	if s.hs != nil {
		if err := s.hs.Stop(); err != nil {
			return err
		}
		s.hs = nil
	}

	if s.crs != nil {
		if err := s.crs.Stop(); err != nil {
			return err
		}
		s.crs = nil
	}

	return nil
}
