package workers

import (
	"log"

	"server/internal/jobs"
	"server/internal/jobs/handlers"
	"server/internal/queue"

	"github.com/hibiken/asynq"
)

var Server *asynq.Server

func Connect() {
	Server = asynq.NewServer(queue.RedisOpt, asynq.Config{
		Concurrency: 10,
	})

	mux := asynq.NewServeMux()

	// Register all task handlers here
	mux.HandleFunc(jobs.JobTypeDocx, handlers.HandleDocx)
	mux.HandleFunc("*", handlers.HandleUnknown)

	if err := Server.Start(mux); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
		return
	}

	log.Println("Successfully started and connected workers")
}
