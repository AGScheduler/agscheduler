package services

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/agscheduler/agscheduler"
	pb "github.com/agscheduler/agscheduler/services/proto"
)

type ClusterProxy struct {
	Scheduler *agscheduler.Scheduler
}

func (c *ClusterProxy) gRPCProxyInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	if !c.Scheduler.IsClusterMode() {
		return handler(ctx, req)
	}

	cn := agscheduler.GetClusterNode(c.Scheduler)
	if cn.IsMainNode() {
		return handler(ctx, req)
	}

	endpointGRPC, ok := cn.MainNode()["endpoint_grpc"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid type for endpoint_grpc")
	}
	conn, err := grpc.NewClient(endpointGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dialing %s failure", endpointGRPC)
	}
	defer conn.Close()

	client := pb.NewSchedulerClient(conn)

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata information")
	}
	if values := md["auth-password-sha2"]; len(values) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, "auth-password-sha2", values[0])
	}

	methodParts := strings.Split(info.FullMethod, "/")
	methodName := methodParts[len(methodParts)-1]
	method := reflect.ValueOf(client).MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method not found: %s", info.FullMethod)
	}

	args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(req)}
	responseValues := method.Call(args)
	resp = responseValues[0].Interface()
	if errorValue := responseValues[1].Interface(); errorValue != nil {
		err = errorValue.(error)
	}
	return resp, err
}
