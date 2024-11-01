package schedulerservice

import (
	"github.com/gilwong00/task-runner/internal/taskdb"
)

type SchedulerService struct {
	Port  int
	Store *taskdb.Store
}

func NewSchedulerService(port int, store *taskdb.Store) *SchedulerService {
	return &SchedulerService{
		Port:  port,
		Store: store,
	}
}

func (s *SchedulerService) Start() error {
	// start http service
	return nil
}
