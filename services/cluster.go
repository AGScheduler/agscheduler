package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kwkwc/agscheduler"
)

type ClusterService struct {
	Cn *agscheduler.ClusterNode
}

func (s *ClusterService) Start() error {
	srservice := &SchedulerRPCService{Scheduler: s.Cn.Scheduler}
	srservice.Address = s.Cn.SchedulerEndpoint
	err := srservice.Start()
	if err != nil {
		return err
	}

	shservice := &SchedulerHTTPService{Scheduler: s.Cn.Scheduler}
	shservice.Address = s.Cn.SchedulerEndpointHTTP
	err = shservice.Start()
	if err != nil {
		return err
	}

	crservice := &clusterRPCService{Cn: s.Cn}
	err = crservice.Start()
	if err != nil {
		return err
	}

	chservice := &clusterHTTPService{Cn: s.Cn}
	err = chservice.Start()
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
