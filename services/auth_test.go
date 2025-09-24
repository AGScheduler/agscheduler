package services

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/agscheduler/agscheduler"
	pb "github.com/agscheduler/agscheduler/services/proto"
	"github.com/agscheduler/agscheduler/stores"
)

var passwordSha2 = "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918"

func testHTTPAuth(t *testing.T, baseUrl string) {
	client := &http.Client{}

	resp, err := http.Get(baseUrl + "/info")
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	req, err := http.NewRequest(http.MethodGet, baseUrl+"/info", bytes.NewReader([]byte{}))
	assert.NoError(t, err)
	req.Header.Add("Auth-Password-SHA2", passwordSha2)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func testGRPCAuth(t *testing.T, c pb.BaseClient) {
	ctx := context.Background()

	_, err := c.GetInfo(ctx, &emptypb.Empty{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no `auth-password-sha2` key")

	ctx = metadata.AppendToOutgoingContext(ctx, "auth-password-sha2", "")
	_, err = c.GetInfo(ctx, &emptypb.Empty{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")

	ctx = context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "auth-password-sha2", passwordSha2)
	_, err = c.GetInfo(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
}

func TestAuth(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	scheduler := &agscheduler.Scheduler{}

	store := &stores.MemoryStore{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	hservice := HTTPService{
		Scheduler:    scheduler,
		PasswordSha2: passwordSha2,
	}
	err = hservice.Start()
	assert.NoError(t, err)

	grservice := GRPCService{
		Scheduler:    scheduler,
		PasswordSha2: passwordSha2,
	}
	err = grservice.Start()
	assert.NoError(t, err)

	conn, err := grpc.NewClient(grservice.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer func() {
		err = conn.Close()
		assert.NoError(t, err)
	}()

	time.Sleep(time.Second)

	baseUrl := "http://" + hservice.Address
	testHTTPAuth(t, baseUrl)
	clientB := pb.NewBaseClient(conn)
	testGRPCAuth(t, clientB)

	err = hservice.Stop()
	assert.NoError(t, err)

	err = grservice.Stop()
	assert.NoError(t, err)

	err = store.Clear()
	assert.NoError(t, err)
}
