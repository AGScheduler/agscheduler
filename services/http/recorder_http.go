package services

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/agscheduler/agscheduler"
)

type req struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type rHTTPService struct {
	recorder *agscheduler.Recorder
}

func (rhs *rHTTPService) handleErr(err error) string {
	if err != nil {
		return err.Error()
	} else {
		return ""
	}
}

func (rhs *rHTTPService) getRecords(c *gin.Context) {
	var r req
	if err := c.ShouldBindQuery(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": rhs.handleErr(err)})
		return
	}

	r.Page = fixPositiveNum(r.Page, 1)
	r.PageSize = fixPositiveNumMax(fixPositiveNum(r.PageSize, 10), 1000)

	var rs []agscheduler.Record
	var total int64
	var err error
	jobId := c.Param("job_id")
	if jobId != "" {
		rs, total, err = rhs.recorder.GetRecords(jobId, r.Page, r.PageSize)
	} else {
		rs, total, err = rhs.recorder.GetAllRecords(r.Page, r.PageSize)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": nil, "error": rhs.handleErr(err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"res":       rs,
			"page":      r.Page,
			"page_size": r.PageSize,
			"total":     total},
		"error": rhs.handleErr(err),
	})
}

func (rhs *rHTTPService) deleteRecords(c *gin.Context) {
	err := rhs.recorder.DeleteRecords(c.Param("job_id"))
	c.JSON(200, gin.H{"data": nil, "error": rhs.handleErr(err)})
}

func (rhs *rHTTPService) deleteAllRecords(c *gin.Context) {
	err := rhs.recorder.DeleteAllRecords()
	c.JSON(200, gin.H{"data": nil, "error": rhs.handleErr(err)})
}

func (rhs *rHTTPService) registerRoutes(r *gin.Engine) {
	r.GET("/recorder/records/:job_id", rhs.getRecords)
	r.GET("/recorder/records", rhs.getRecords)
	r.DELETE("/recorder/records/:job_id", rhs.deleteRecords)
	r.DELETE("/recorder/records", rhs.deleteAllRecords)
}

func fixPositiveNum(num, numDef int) int {
	if num < 1 {
		return numDef
	}

	return num
}

func fixPositiveNumMax(num, numMax int) int {
	if num > numMax {
		return numMax
	}

	return num
}
