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

	if s.Cn.MainEndpoint != s.Cn.Endpoint {
		err = s.Cn.RegisterNodeRemote(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to register node remote: %s", err)
		}
	}

	return nil
}
