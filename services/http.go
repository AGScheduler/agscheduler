package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/kwkwc/agscheduler"
)

type HTTPService struct {
	Scheduler *agscheduler.Scheduler

	// Default: `127.0.0.1:36370`
	Address string

	srv *http.Server
}

func (s *HTTPService) Start() error {
	if s.Address == "" {
		s.Address = "127.0.0.1:36370"
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())

	chap := &ClusterProxy{Scheduler: s.Scheduler}
	r.Use(chap.GinProxy())

	shs := &sHTTPService{scheduler: s.Scheduler}
	shs.registerRoutes(r)

	if s.Scheduler.IsClusterMode() {
		chs := &cHTTPService{cn: agscheduler.GetClusterNode(s.Scheduler)}
		chs.registerRoutes(r)
	}

	slog.Info(fmt.Sprintf("HTTP Service listening at: %s", s.Address))

	s.srv = &http.Server{
		Addr:    s.Address,
		Handler: r,
	}

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error(fmt.Sprintf("HTTP Service Unavailable: %s", err))
		}
	}()

	return nil
}

func (s *HTTPService) Stop() error {
	slog.Info("HTTP Service stop")

	if err := s.srv.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("failed to stop service: %s", err)
	}

	return nil
}
