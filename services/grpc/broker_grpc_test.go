package services

import (
	"context"
	"testing"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/stretchr/testify/assert"

	pb "github.com/agscheduler/agscheduler/services/proto"
)

func testBrokerGRPC(t *testing.T, c pb.BrokerClient) {
	ctx := context.Background()

	qsResp, err := c.GetQueues(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Len(t, qsResp.Queues, 1)
}
