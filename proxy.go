package agscheduler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/kwkwc/agscheduler/services/proto"
)

type ClusterHAProxy struct {
	Scheduler *Scheduler
}

func (c *ClusterHAProxy) GinProxy() gin.HandlerFunc {
	return func(gc *gin.Context) {
		if !c.Scheduler.IsClusterMode() {
			return
		}

		if c.Scheduler.clusterNode.IsMainNode() {
			return
		}

		proxyUrl := new(url.URL)
		if gc.Request.TLS == nil {
			proxyUrl.Scheme = "http"
		} else {
			proxyUrl.Scheme = "https"
		}

		schedulerEndpointHTTP, ok := c.Scheduler.clusterNode.MainNode()["scheduler_endpoint_http"].(string)
		if !ok {
			gc.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type for scheduler_endpoint_http"})
			gc.Abort()
		}
		proxyUrl.Host = schedulerEndpointHTTP

		proxy := httputil.NewSingleHostReverseProxy(proxyUrl)
		proxy.ServeHTTP(gc.Writer, gc.Request)
	}
}

func (c *ClusterHAProxy) GRPCProxyInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	if !c.Scheduler.IsClusterMode() {
		return handler(ctx, req)
	}

	cn := GetClusterNode(c.Scheduler)
	if cn.IsMainNode() {
		return handler(ctx, req)
	}

	schedulerEndpoint, ok := cn.MainNode()["scheduler_endpoint"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid type for scheduler_endpoint")
	}
	conn, err := grpc.Dial(schedulerEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dialing %s failure", schedulerEndpoint)
	}
	defer conn.Close()

	client := pb.NewSchedulerClient(conn)

	methodParts := strings.Split(info.FullMethod, "/")
	methodName := methodParts[len(methodParts)-1]
	method := reflect.ValueOf(client).MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method not found: %s", info.FullMethod)
	}

	args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(req)}
	responseValues := method.Call(args)
	resp = responseValues[0].Interface()
	errInter := responseValues[1].Interface()
	if errInter != nil {
		err = errInter.(error)
	}

	return resp, err
}
