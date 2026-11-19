package services

import (
	"context"
	"testing"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/stretchr/testify/assert"

	pb "github.com/agscheduler/agscheduler/services/proto"
)

func testClusterGRPC(t *testing.T, c pb.ClusterClient) {
	ctx := context.Background()

	nsResp, err := c.GetNodes(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Len(t, nsResp.Nodes, 2)
}
