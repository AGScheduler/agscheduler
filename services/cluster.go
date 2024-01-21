package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kwkwc/agscheduler"
)

type ClusterService struct {
	Cn *agscheduler.ClusterNode

	srs *SchedulerGRPCService
	shs *SchedulerHTTPService
	crs *clusterRPCService
	chs *clusterHTTPService
}

func (s *ClusterService) Start() error {
	s.srs = &SchedulerGRPCService{Scheduler: s.Cn.Scheduler}
	s.srs.Address = s.Cn.SchedulerEndpoint
	err := s.srs.Start()
	if err != nil {
		return err
	}

	s.shs = &SchedulerHTTPService{Scheduler: s.Cn.Scheduler}
	s.shs.Address = s.Cn.SchedulerEndpointHTTP
	err = s.shs.Start()
	if err != nil {
		return err
	}

	s.crs = &clusterRPCService{Cn: s.Cn}
	err = s.crs.Start()
	if err != nil {
		return err
	}

	s.chs = &clusterHTTPService{Cn: s.Cn}
	err = s.chs.Start()
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Cluster Queue: `%s`", s.Cn.Queue))

	if !s.Cn.IsMainNode() {
		err = s.Cn.RegisterNodeRemote(context.TODO())
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to register node remote: %s\n", err))
		}
	}

	return nil
}

func (s *ClusterService) Stop() error {
	err := s.srs.Stop()
	if err != nil {
		return err
	}

	err = s.shs.Stop()
	if err != nil {
		return err
	}

	err = s.crs.Stop()
	if err != nil {
		return err
	}

	err = s.chs.Stop()
	if err != nil {
		return err
	}

	return nil
}
