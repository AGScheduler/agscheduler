package services

import (
	"fmt"
	"log/slog"

	"github.com/kwkwc/agscheduler"
)

type ClusterService struct {
	Scheduler *agscheduler.Scheduler
	Cn        *agscheduler.ClusterNode
}

func (s *ClusterService) Start() error {
	if s.Cn.Endpoint == "" {
		s.Cn.Endpoint = "127.0.0.1:36364"
	}
	if s.Cn.MainEndpoint == "" {
		s.Cn.MainEndpoint = s.Cn.Endpoint
	}
	if s.Cn.SchedulerEndpoint == "" {
		s.Cn.SchedulerEndpoint = "127.0.0.1:36363"
	}
	if s.Cn.Queue == "" {
		s.Cn.Queue = "default"
	}

	srservice := &SchedulerRPCService{Scheduler: s.Scheduler}
	srservice.Address = s.Cn.SchedulerEndpoint
	err := srservice.Start()
	if err != nil {
		return err
	}

	crservice := &clusterRPCService{Scheduler: s.Scheduler, Cn: s.Cn}
	err = crservice.Start()
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Cluster Node Queue: `%s`", s.Cn.Queue))

	return nil
}
