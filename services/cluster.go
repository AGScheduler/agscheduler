package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/agscheduler/agscheduler"
)

type ClusterService struct {
	Cn *agscheduler.ClusterNode

	grs *GRPCService
	hs  *HTTPService
	crs *clusterRPCService
}

func (s *ClusterService) Start() error {
	s.grs = &GRPCService{Scheduler: s.Cn.Scheduler}
	s.grs.Address = s.Cn.EndpointGRPC
	err := s.grs.Start()
	if err != nil {
		return err
	}

	s.hs = &HTTPService{Scheduler: s.Cn.Scheduler}
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
			slog.Error(fmt.Sprintf("Failed to register node remote: %s\n", err))
		}
	}

	return nil
}

func (s *ClusterService) Stop() error {
	err := s.grs.Stop()
	if err != nil {
		return err
	}

	err = s.hs.Stop()
	if err != nil {
		return err
	}

	err = s.crs.Stop()
	if err != nil {
		return err
	}

	return nil
}
