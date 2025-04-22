package coordinatorservice

import (
	"context"

	pb "github.com/gilwong00/task-runner/proto/gen"
	"github.com/gofrs/uuid/v5"
)

func (c *CoordinatorServer) SubmitTask(
	ctx context.Context,
	req *pb.SubmitTaskRequest,
) (*pb.SubmitTaskResponse, error) {
	newUUID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	taskId := newUUID.String()
	receiveTaskRequest := &pb.ReceiveTaskRequest{
		TaskId:  taskId,
		Payload: req.Payload,
	}
	if err := c.submitTaskToWorker(receiveTaskRequest); err != nil {
		return nil, err
	}
	return &pb.SubmitTaskResponse{
		TaskId:  taskId,
		Message: "Task submitted successfully",
	}, nil
}
