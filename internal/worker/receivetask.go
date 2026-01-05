package workerservice

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	workerv1 "github.com/gilwong00/task-runner/internal/gen/worker/v1"
)

func (s *workerService) ReceiveTask(
	ctx context.Context,
	req *connect.Request[workerv1.ReceiveTaskRequest],
) (*connect.Response[workerv1.ReceiveTaskResponse], error) {
	return nil, errors.New("not implemented")
}
