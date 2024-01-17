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

type cHTTPService struct {
	cn *agscheduler.ClusterNode
}

func (chs *cHTTPService) nodes(c *gin.Context) {
	c.JSON(200, gin.H{"data": chs.cn.NodeMap(), "error": ""})
}

type clusterHTTPService struct {
	Cn *agscheduler.ClusterNode

	srv *http.Server
}

func (s *clusterHTTPService) registerRoutes(r *gin.Engine, shs *cHTTPService) {
	r.GET("/cluster/nodes", shs.nodes)
}

func (s *clusterHTTPService) Start() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())

	s.registerRoutes(r, &cHTTPService{cn: s.Cn})

	slog.Info(fmt.Sprintf("cluster HTTP Service listening at: %s", s.Cn.EndpointHTTP))

	s.srv = &http.Server{
		Addr:    s.Cn.EndpointHTTP,
		Handler: r,
	}

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error(fmt.Sprintf("Cluster HTTP Service Unavailable: %s", err))
		}
	}()

	return nil
}

func (s *clusterHTTPService) Stop() error {
	slog.Info("Cluster HTTP Service stop")

	if err := s.srv.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("failed to stop service: %s", err)
	}

	return nil
}
