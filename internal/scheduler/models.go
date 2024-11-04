package schedulerservice

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
	Command string `json:"command"`
	// ISO 8601 format
	// example: 2006-01-02T15:04:05+07:00 or 2024-11-03T20:02:03Z
	ScheduledAt string `json:"scheduledAt"`
}
