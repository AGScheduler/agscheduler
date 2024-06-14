package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/agscheduler/agscheduler/services/proto"
)

func testBrokerGRPC(t *testing.T, c pb.BrokerClient) {
	ctx := context.Background()

	qsResp, err := c.GetQueues(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Len(t, qsResp.Queues, 1)
}
