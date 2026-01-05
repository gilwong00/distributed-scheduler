package coordinatorservice

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	coordinatorv1 "github.com/gilwong00/task-runner/internal/gen/proto/coordinator/v1"
)

func (c *coordinatorService) UpdateTaskStatus(
	ctx context.Context,
	req *connect.Request[coordinatorv1.UpdateTaskStatusRequest],
) (*connect.Response[coordinatorv1.UpdateTaskStatusResponse], error) {
	return nil, errors.New("not implemented")
}
