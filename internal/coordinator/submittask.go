package coordinatorservice

import (
	"context"
	"errors"

	pb "github.com/gilwong00/task-runner/proto/gen"
)

func (s *CoordinatorServer) SubmitTask(
	ctx context.Context,
	req *pb.SubmitTaskRequest,
) (*pb.SubmitTaskResponse, error) {
	return nil, errors.New("not implemented")
}
