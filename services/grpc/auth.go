package services

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
)

func gRPCVerifyPassword(ctx context.Context, passwordSha2 string) error {
	if passwordSha2 == "" {
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("no metadata information")
	}
	values, ok := md["auth-password-sha2"]
	if !ok || len(values) == 0 {
		return fmt.Errorf("no `auth-password-sha2` key")
	}
	if passwordSha2 != values[0] {
		return fmt.Errorf("unauthorized")
	}
	return nil
}
