package coordinatorservice

import (
	"context"
	"errors"

	pb "github.com/gilwong00/task-runner/proto/gen"
)

func (c *CoordinatorServer) UpdateTaskStatus(
	ctx context.Context,
	req *pb.UpdateTaskStatusRequest,
) (*pb.UpdateTaskStatusResponse, error) {
	return nil, errors.New("not implemented")
}
