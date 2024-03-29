package services

import (
	"github.com/gin-gonic/gin"

	"github.com/agscheduler/agscheduler"
)

type cHTTPService struct {
	cn *agscheduler.ClusterNode
}

func (chs *cHTTPService) nodes(c *gin.Context) {
	c.JSON(200, gin.H{"data": chs.cn.NodeMapCopy(), "error": ""})
}

func (chs *cHTTPService) registerRoutes(r *gin.Engine) {
	r.GET("/cluster/nodes", chs.nodes)
}
