package services

import (
	"github.com/gin-gonic/gin"

	"github.com/agscheduler/agscheduler"
)

type brkHTTPService struct {
	broker *agscheduler.Broker
}

func (bhs *brkHTTPService) getQueues(c *gin.Context) {
	qs := bhs.broker.GetQueues()
	c.JSON(200, gin.H{"data": qs, "error": ""})
}

func (bhs *brkHTTPService) registerRoutes(r *gin.Engine) {
	r.GET("/broker/queues", bhs.getQueues)
}
