package agscheduler

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func ClusterHAGinProxy(s *Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.IsClusterMode() {
			return
		}

		if s.clusterNode.IsMainNode() {
			return
		}

		proxyUrl := new(url.URL)
		if c.Request.TLS == nil {
			proxyUrl.Scheme = "http"
		} else {
			proxyUrl.Scheme = "https"
		}
		proxyUrl.Host = s.clusterNode.MainNode()["scheduler_endpoint_http"].(string)
		c.Request.Host = proxyUrl.Host

		proxy := httputil.NewSingleHostReverseProxy(proxyUrl)

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
