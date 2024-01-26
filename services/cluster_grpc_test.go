package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/kwkwc/agscheduler/services/proto"
)

func testClusterGRPC(t *testing.T, c pb.ClusterClient) {
	ctx := context.Background()

	pbNs, err := c.GetNodes(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Len(t, pbNs.Nodes, 2)
}
