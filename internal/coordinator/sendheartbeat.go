package coordinatorservice

import (
	"context"
	"errors"

	pb "github.com/gilwong00/task-runner/proto/gen"
)

func (c *CoordinatorServer) SendHeartbeat(
	ctx context.Context,
	req *pb.SendHeartbeatRequest,
) (*pb.SendHeartbeatResponse, error) {
	return nil, errors.New("not implemented")
}
