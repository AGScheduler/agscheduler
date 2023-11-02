package services

import (
	"fmt"
	"log/slog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/kwkwc/agscheduler"
)

type sHTTPService struct {
	scheduler *agscheduler.Scheduler
}

func (shs *sHTTPService) handleJob(j agscheduler.Job, err error) gin.H {
	if j.Id == "" {
		return gin.H{"data": nil, "error": shs.handleErr(err)}
	} else {
		return gin.H{"data": j, "error": shs.handleErr(err)}
	}
}

func (shs *sHTTPService) handleErr(err error) string {
	if err != nil {
		return err.Error()
	} else {
		return ""
	}
}

func (shs *sHTTPService) AddJob(c *gin.Context) {
	j := agscheduler.Job{}
	err := c.BindJSON(&j)
	if err != nil {
		c.JSON(400, shs.handleJob(j, err))
		return
	}

	j, err = shs.scheduler.AddJob(j)
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) GetJob(c *gin.Context) {
	j, err := shs.scheduler.GetJob(c.Param("id"))
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) GetAllJobs(c *gin.Context) {
	js, err := shs.scheduler.GetAllJobs()
	c.JSON(200, gin.H{"data": js, "error": shs.handleErr(err)})
}

func (shs *sHTTPService) UpdateJob(c *gin.Context) {
	j := agscheduler.Job{}
	err := c.BindJSON(&j)
	if err != nil {
		c.JSON(400, shs.handleJob(j, err))
		return
	}

	j, err = shs.scheduler.UpdateJob(j)
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) DeleteJob(c *gin.Context) {
	err := shs.scheduler.DeleteJob(c.Param("id"))
	c.JSON(200, gin.H{"data": nil, "error": shs.handleErr(err)})
}

func (shs *sHTTPService) DeleteAllJobs(c *gin.Context) {
	shs.scheduler.DeleteAllJobs()
	c.JSON(200, gin.H{"data": nil, "error": ""})
}

func (shs *sHTTPService) PauseJob(c *gin.Context) {
	j, err := shs.scheduler.PauseJob(c.Param("id"))
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) ResumeJob(c *gin.Context) {
	j, err := shs.scheduler.ResumeJob(c.Param("id"))
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) RunJob(c *gin.Context) {
	j := agscheduler.Job{}
	err := c.BindJSON(&j)
	if err != nil {
		c.JSON(400, shs.handleJob(j, err))
		return
	}

	err = shs.scheduler.RunJob(j)
	c.JSON(200, gin.H{"data": nil, "error": shs.handleErr(err)})
}

func (shs *sHTTPService) Start(c *gin.Context) {
	shs.scheduler.Start()
	c.JSON(200, gin.H{"data": nil, "error": ""})
}

func (shs *sHTTPService) Stop(c *gin.Context) {
	shs.scheduler.Stop()
	c.JSON(200, gin.H{"data": nil, "error": ""})
}

type SchedulerHTTPService struct {
	Scheduler *agscheduler.Scheduler
	Address   string
}

func (s *SchedulerHTTPService) registerRoutes(r *gin.Engine, shs *sHTTPService) {
	r.POST("/scheduler/job", shs.AddJob)
	r.GET("/scheduler/job/:id", shs.GetJob)
	r.GET("/scheduler/jobs", shs.GetAllJobs)
	r.PUT("/scheduler/job", shs.UpdateJob)
	r.DELETE("/scheduler/job/:id", shs.DeleteJob)
	r.DELETE("/scheduler/jobs", shs.DeleteAllJobs)
	r.POST("/scheduler/job/:id/pause", shs.PauseJob)
	r.POST("/scheduler/job/:id/resume", shs.ResumeJob)
	r.POST("/scheduler/job/run", shs.RunJob)
	r.POST("/scheduler/start", shs.Start)
	r.POST("/scheduler/stop", shs.Stop)
}

func (s *SchedulerHTTPService) Start() error {
	if s.Address == "" {
		s.Address = "127.0.0.1:63636"
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())

	s.registerRoutes(r, &sHTTPService{scheduler: s.Scheduler})

	slog.Info(fmt.Sprintf("Scheduler HTTP Service listening at: %s", s.Address))

	go func() {
		if err := r.Run(s.Address); err != nil {
			slog.Error(fmt.Sprintf("Scheduler HTTP Service Unavailable: %s", err))
		}
	}()

	return nil
}
