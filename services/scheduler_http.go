package services

import (
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

func (shs *sHTTPService) addJob(c *gin.Context) {
	j := agscheduler.Job{}
	err := c.BindJSON(&j)
	if err != nil {
		c.JSON(400, shs.handleJob(j, err))
		return
	}

	j, err = shs.scheduler.AddJob(j)
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) getJob(c *gin.Context) {
	j, err := shs.scheduler.GetJob(c.Param("id"))
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) getAllJobs(c *gin.Context) {
	js, err := shs.scheduler.GetAllJobs()
	c.JSON(200, gin.H{"data": js, "error": shs.handleErr(err)})
}

func (shs *sHTTPService) updateJob(c *gin.Context) {
	j := agscheduler.Job{}
	err := c.BindJSON(&j)
	if err != nil {
		c.JSON(400, shs.handleJob(j, err))
		return
	}

	j, err = shs.scheduler.UpdateJob(j)
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) deleteJob(c *gin.Context) {
	err := shs.scheduler.DeleteJob(c.Param("id"))
	c.JSON(200, gin.H{"data": nil, "error": shs.handleErr(err)})
}

func (shs *sHTTPService) deleteAllJobs(c *gin.Context) {
	err := shs.scheduler.DeleteAllJobs()
	c.JSON(200, gin.H{"data": nil, "error": shs.handleErr(err)})
}

func (shs *sHTTPService) pauseJob(c *gin.Context) {
	j, err := shs.scheduler.PauseJob(c.Param("id"))
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) resumeJob(c *gin.Context) {
	j, err := shs.scheduler.ResumeJob(c.Param("id"))
	c.JSON(200, shs.handleJob(j, err))
}

func (shs *sHTTPService) runJob(c *gin.Context) {
	j := agscheduler.Job{}
	err := c.BindJSON(&j)
	if err != nil {
		c.JSON(400, shs.handleJob(j, err))
		return
	}

	err = shs.scheduler.RunJob(j)
	c.JSON(200, gin.H{"data": nil, "error": shs.handleErr(err)})
}

func (shs *sHTTPService) scheduleJob(c *gin.Context) {
	j := agscheduler.Job{}
	err := c.BindJSON(&j)
	if err != nil {
		c.JSON(400, shs.handleJob(j, err))
		return
	}

	err = shs.scheduler.ScheduleJob(j)
	c.JSON(200, gin.H{"data": nil, "error": shs.handleErr(err)})
}

func (shs *sHTTPService) start(c *gin.Context) {
	shs.scheduler.Start()
	c.JSON(200, gin.H{"data": nil, "error": ""})
}

func (shs *sHTTPService) stop(c *gin.Context) {
	shs.scheduler.Stop()
	c.JSON(200, gin.H{"data": nil, "error": ""})
}

func (shs *sHTTPService) registerRoutes(r *gin.Engine) {
	r.POST("/scheduler/job", shs.addJob)
	r.GET("/scheduler/job/:id", shs.getJob)
	r.GET("/scheduler/jobs", shs.getAllJobs)
	r.PUT("/scheduler/job", shs.updateJob)
	r.DELETE("/scheduler/job/:id", shs.deleteJob)
	r.DELETE("/scheduler/jobs", shs.deleteAllJobs)
	r.POST("/scheduler/job/:id/pause", shs.pauseJob)
	r.POST("/scheduler/job/:id/resume", shs.resumeJob)
	r.POST("/scheduler/job/run", shs.runJob)
	r.POST("/scheduler/job/schedule", shs.scheduleJob)
	r.POST("/scheduler/start", shs.start)
	r.POST("/scheduler/stop", shs.stop)
}
