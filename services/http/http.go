package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/agscheduler/agscheduler"
)

type bHTTPService struct {
	scheduler *agscheduler.Scheduler
}

func (bhs *bHTTPService) info(c *gin.Context) {
	c.JSON(200, gin.H{"data": bhs.scheduler.Info(), "error": ""})
}

func (bhs *bHTTPService) funcs(c *gin.Context) {
	c.JSON(200, gin.H{"data": agscheduler.FuncMapReadable(), "error": ""})
}

func (bhs *bHTTPService) registerRoutes(r *gin.Engine) {
	r.GET("/info", bhs.info)
	r.GET("/funcs", bhs.funcs)
}

type HTTPService struct {
	Scheduler *agscheduler.Scheduler

	// Default: `127.0.0.1:36370`
	Address string
	// SHA256 encrypted authorization password, e.g. here is admin
	// echo -n admin | shasum -a 256
	// `8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918`
	//
	// Add `Auth-Password-SHA2` header on request.
	PasswordSha2 string

	srv *http.Server
}

func (s *HTTPService) Start() error {
	if s.Address == "" {
		s.Address = "127.0.0.1:36370"
	}

	r := gin.Default()
	r.Use(ginCors())
	r.Use(ginVerifyPassword(s.PasswordSha2))

	cp := &ClusterProxy{Scheduler: s.Scheduler}
	r.Use(cp.ginProxy())

	bhs := &bHTTPService{scheduler: s.Scheduler}
	bhs.registerRoutes(r)

	shs := &sHTTPService{scheduler: s.Scheduler}
	shs.registerRoutes(r)

	if s.Scheduler.HasBroker() {
		bhs := &brkHTTPService{broker: agscheduler.GetBroker(s.Scheduler)}
		bhs.registerRoutes(r)
	}

	if s.Scheduler.HasRecorder() {
		rhs := &rHTTPService{recorder: agscheduler.GetRecorder(s.Scheduler)}
		rhs.registerRoutes(r)
	}

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
