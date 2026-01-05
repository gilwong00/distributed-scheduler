package coordinatorservice

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	coordinatorv1 "github.com/gilwong00/task-runner/internal/gen/proto/coordinator/v1"
)

func (c *coordinatorService) ReceiveHeartbeat(
	ctx context.Context,
	req *connect.Request[coordinatorv1.ReceiveHeartbeatRequest],
) (*connect.Response[coordinatorv1.ReceiveHeartbeatResponse], error) {
	c.WorkerPoolMutex.Lock()
	defer c.WorkerPoolMutex.Unlock()

	return nil, errors.New("not implemented")
}
