package workers

import (
	"log"

	"server/internal/config"
	"server/internal/jobs/handlers"

	"github.com/hibiken/asynq"
)

var Server *asynq.Server

func Connect() {
	Server = asynq.NewServer(config.RedisOpt, asynq.Config{
		Concurrency: 10,
	})

	mux := asynq.NewServeMux()

	// Register all task handlers here
	// TODO: uncomment in next PR
	// mux.HandleFunc(jobs.JobTypeDocx, handlers.HandleDocx)
	mux.HandleFunc("*", handlers.HandleUnknown)

	if err := Server.Start(mux); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
		return
	}

	log.Println("Successfully started and connected workers")
}
