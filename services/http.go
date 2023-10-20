package services

import (
	"fmt"
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/kwkwc/agscheduler"
)

type HTTPService struct {
	scheduler *agscheduler.Scheduler
}

func (hs *HTTPService) handleJob(j agscheduler.Job, err error) gin.H {
	if j.Id == "" {
		return gin.H{"data": nil, "error": hs.handleErr(err)}
	} else {
		return gin.H{"data": j, "error": hs.handleErr(err)}
	}
}

func (hs *HTTPService) handleErr(err error) string {
	if err != nil {
		return err.Error()
	} else {
		return ""
	}
}

func (hs *HTTPService) AddJob(c *gin.Context) {
	j := agscheduler.Job{}
	err := c.BindJSON(&j)
	if err != nil {
		c.JSON(400, hs.handleJob(j, err))
		return
	}

	j, err = hs.scheduler.AddJob(j)
	c.JSON(200, hs.handleJob(j, err))
}

func (hs *HTTPService) GetJob(c *gin.Context) {
	j, err := hs.scheduler.GetJob(c.Param("id"))
	c.JSON(200, hs.handleJob(j, err))
}

func (hs *HTTPService) GetAllJobs(c *gin.Context) {
	js, err := hs.scheduler.GetAllJobs()
	c.JSON(200, gin.H{"data": js, "error": hs.handleErr(err)})
}

func (hs *HTTPService) UpdateJob(c *gin.Context) {
	j := agscheduler.Job{}
	err := c.BindJSON(&j)
	if err != nil {
		c.JSON(400, hs.handleJob(j, err))
		return
	}

	j, err = hs.scheduler.UpdateJob(j)
	c.JSON(200, hs.handleJob(j, err))
}

func (hs *HTTPService) DeleteJob(c *gin.Context) {
	err := hs.scheduler.DeleteJob(c.Param("id"))
	c.JSON(200, gin.H{"data": nil, "error": hs.handleErr(err)})
}

func (hs *HTTPService) DeleteAllJobs(c *gin.Context) {
	hs.scheduler.DeleteAllJobs()
	c.JSON(200, gin.H{"data": nil, "error": ""})
}

func (hs *HTTPService) PauseJob(c *gin.Context) {
	j, err := hs.scheduler.PauseJob(c.Param("id"))
	c.JSON(200, hs.handleJob(j, err))
}

func (hs *HTTPService) ResumeJob(c *gin.Context) {
	j, err := hs.scheduler.ResumeJob(c.Param("id"))
	c.JSON(200, hs.handleJob(j, err))
}

func (hs *HTTPService) Start(c *gin.Context) {
	hs.scheduler.Start()
	c.JSON(200, gin.H{"data": nil, "error": ""})
}

func (hs *HTTPService) Stop(c *gin.Context) {
	hs.scheduler.Stop()
	c.JSON(200, gin.H{"data": nil, "error": ""})
}

type SchedulerHTTPService struct {
	Scheduler *agscheduler.Scheduler
}

func (s *SchedulerHTTPService) registerRoutes(r *gin.Engine, hs *HTTPService) {
	r.POST("/scheduler/job", hs.AddJob)
	r.GET("/scheduler/job/:id", hs.GetJob)
	r.GET("/scheduler/jobs", hs.GetAllJobs)
	r.PUT("/scheduler/job", hs.UpdateJob)
	r.DELETE("/scheduler/job/:id", hs.DeleteJob)
	r.DELETE("/scheduler/jobs", hs.DeleteAllJobs)
	r.POST("/scheduler/job/:id/pause", hs.PauseJob)
	r.POST("/scheduler/job/:id/resume", hs.ResumeJob)
	r.POST("/scheduler/start", hs.Start)
	r.POST("/scheduler/stop", hs.Stop)
}

func (s *SchedulerHTTPService) Start(address string) error {
	if address == "" {
		address = "127.0.0.1:63636"
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())

	s.registerRoutes(r, &HTTPService{scheduler: s.Scheduler})

	slog.Info(fmt.Sprintf("Scheduler HTTP Service listening at: %s", address))

	go func() {
		if err := r.Run(address); err != nil {
			slog.Error(fmt.Sprintf("Scheduler HTTP Service Unavailable: %s", err))
		}
	}()

	return nil
}
