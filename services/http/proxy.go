package services

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/agscheduler/agscheduler"
)

type ClusterProxy struct {
	Scheduler *agscheduler.Scheduler
}

func (c *ClusterProxy) ginProxy() gin.HandlerFunc {
	return func(gc *gin.Context) {
		if !c.Scheduler.IsClusterMode() {
			return
		}

		cn := agscheduler.GetClusterNode(c.Scheduler)
		if cn.IsMainNode() {
			return
		}

		proxyURL := new(url.URL)
		if gc.Request.TLS == nil {
			proxyURL.Scheme = "http"
		} else {
			proxyURL.Scheme = "https"
		}

		endpointHTTP, ok := cn.MainNode()["endpoint_http"].(string)
		if !ok {
			gc.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type for endpoint_http"})
			gc.Abort()
			return
		}
		proxyURL.Host = endpointHTTP

		proxy := httputil.NewSingleHostReverseProxy(proxyURL)
		proxy.ModifyResponse = func(resp *http.Response) error {
			resp.Header.Del("Access-Control-Allow-Origin")
			return nil
		}
		proxy.ServeHTTP(gc.Writer, gc.Request)
		gc.Abort()
	}
}
