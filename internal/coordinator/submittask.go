package coordinatorservice

import (
	"context"

	"connectrpc.com/connect"
	coordinatorv1 "github.com/gilwong00/task-runner/internal/gen/coordinator/v1"
	workerv1 "github.com/gilwong00/task-runner/internal/gen/worker/v1"

	"github.com/gofrs/uuid/v5"
)

func (c *coordinatorService) SubmitTask(
	ctx context.Context,
	req *connect.Request[coordinatorv1.SubmitTaskRequest],
) (*connect.Response[coordinatorv1.SubmitTaskResponse], error) {
	newUUID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	taskId := newUUID.String()
	receiveTaskRequest := &workerv1.ReceiveTaskRequest{
		TaskId:  taskId,
		Payload: req.Msg.Payload,
	}
	if err := c.submitTaskToWorker(receiveTaskRequest); err != nil {
		return nil, err
	}
	return connect.NewResponse(&coordinatorv1.SubmitTaskResponse{
		TaskId:  taskId,
		Message: "Task submitted successfully",
	}), nil
}
