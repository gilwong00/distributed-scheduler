package coordinatorservice

import (
	"context"
	"errors"

	pb "github.com/gilwong00/task-runner/proto/gen"
)

func (c *CoordinatorServer) ReceiveHeartbeat(
	ctx context.Context,
	req *pb.ReceiveHeartbeatRequest,
) (*pb.ReceiveHeartbeatResponse, error) {
	c.WorkerPoolMutex.Lock()
	defer c.WorkerPoolMutex.Unlock()

	return nil, errors.New("not implemented")
}
