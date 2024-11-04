package schedulerservice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gilwong00/task-runner/internal/taskdb"
)

type SchedulerService struct {
	Port  int
	store *taskdb.Store
}

// TODO: move this to a models.go file possibly
type TaskResponse struct {
	TaskID      string `json:"taskId"`
	Command     string `json:"command"`
	ScheduledAt string `json:"scheduledAt,omitempty"`
	PickedAt    string `json:"pickedAt,omitempty"`
	StartedAt   string `json:"startedAt,omitempty"`
	CompletedAt string `json:"completedAt,omitempty"`
	FailedAt    string `json:"failedAt,omitempty"`
}

type ScheduleTaskPayload struct {
	Command     string `json:"command"`
	ScheduledAt string `json:"scheduledAt"` // ISO 8601 format

}

func NewSchedulerService(port int, store *taskdb.Store) *SchedulerService {
	return &SchedulerService{
		Port:  port,
		store: store,
	}
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, errMessage string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(errMessage))
}

func ParseJSONBody[T any](r *http.Request) (T, error) {
	var parsed T
	err := json.NewDecoder(r.Body).Decode(&parsed)
	if err != nil {
		return parsed, fmt.Errorf("unable to parse JSON: `%s`", err)
	}
	return parsed, nil
}

func WriteSuccessResponse(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
