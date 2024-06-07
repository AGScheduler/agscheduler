package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

func ginVerifyPassword(passwordSha2 string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authPasswordSha2 := c.Request.Header.Get("Auth-Password-SHA2")
		if passwordSha2 != "" && passwordSha2 != authPasswordSha2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
		}
	}
}

func gRPCVerifyPassword(ctx context.Context, passwordSha2 string) (err error) {
	if passwordSha2 != "" {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return fmt.Errorf("no metadata information")
		}
		vals, ok := md["auth-password-sha2"]
		if !ok {
			return fmt.Errorf("no `auth-password-sha2` key")
		}
		authPasswordSha2 := vals[0]

		if passwordSha2 != authPasswordSha2 {
			return fmt.Errorf("unauthorized")
		}
	}

	return
}
