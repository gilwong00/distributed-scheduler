package schedulerservice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	taskpostgres "github.com/gilwong00/task-runner/internal/taskdb/gen"
	"github.com/gofrs/uuid/v5"
)

func (s *SchedulerService) StartHttpServer() error {
	// start http service
	mux := http.NewServeMux()
	mux.HandleFunc("POST /task", s.createPost)
	mux.HandleFunc("GET /status/{taskID}", s.getStatus)

	server := http.Server{
		Addr:         fmt.Sprintf(":%v", s.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	// start the server
	go func() {
		fmt.Printf("Starting server on port %v\n", s.Port)
		err := server.ListenAndServe()
		if err != nil {
			fmt.Printf("Error starting server: %s", err.Error())
			os.Exit(1)
		}
	}()
	// trap sigterm or interupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	// Block until a signal is received.
	sig := <-c
	log.Println("Got signal:", sig)
	// gracefully shutdown the server, waiting max 30 seconds for current operations to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func (s *SchedulerService) createPost(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	defer r.Body.Close()
	payload, err := ParseJSONBody[ScheduleTaskPayload](r)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("Received schedule request: %+v", payload)
	// Parse the scheduled at time
	scheduledTime, err := time.Parse(time.RFC3339, payload.ScheduledAt)
	if err != nil {
		http.Error(w, "Invalid date format. Use ISO 8601 format.", http.StatusBadRequest)
		return
	}
	// Convert the scheduled time to Unix timestamp
	unixTimestamp := time.Unix(scheduledTime.Unix(), 0)
	var pgTask taskpostgres.Task
	if err := s.store.ReadWriteTx(ctx, func(tx *taskpostgres.Queries) error {
		var err error
		pgTask, err = tx.CreateTask(ctx, taskpostgres.CreateTaskParams{
			Command:     payload.Command,
			ScheduledAt: unixTimestamp,
		})
		return err
	}); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	response := TaskResponse{
		TaskID:      pgTask.ID.String(),
		Command:     pgTask.Command,
		ScheduledAt: pgTask.ScheduledAt.GoString(),
	}
	WriteSuccessResponse(w, response)
}

func (s *SchedulerService) getStatus(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	taskID := r.PathValue("taskID")
	if taskID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "taskID is required")
		return
	}
	taskUUID, err := uuid.FromString(taskID)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("Looking up status for task with id: %s", taskUUID)
	// get task from DB
	var pgTask taskpostgres.Task
	if err := s.store.ReadTx(ctx, func(tx *taskpostgres.Queries) error {
		var err error
		pgTask, err = tx.GetTaskByID(ctx, taskUUID)
		return err
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			WriteErrorResponse(w, http.StatusNotFound, fmt.Sprintf("failed to find task with id: %s", taskUUID))
			return
		}
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	response := TaskResponse{
		TaskID:      pgTask.ID.String(),
		Command:     pgTask.Command,
		ScheduledAt: "",
		PickedAt:    "",
		StartedAt:   "",
		CompletedAt: "",
		FailedAt:    "",
	}
	// TODO fill out dates before writing response
	WriteSuccessResponse(w, response)
}
