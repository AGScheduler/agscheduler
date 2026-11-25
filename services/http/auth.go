package services

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func ginCors() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowHeaders = append(config.AllowHeaders, "Auth-Password-SHA2")
	config.AllowAllOrigins = true

	return cors.New(config)
}

func ginVerifyPassword(passwordSha2 string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authPasswordSha2 := c.Request.Header.Get("Auth-Password-SHA2")
		if passwordSha2 != "" && passwordSha2 != authPasswordSha2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
		}
	}
}
