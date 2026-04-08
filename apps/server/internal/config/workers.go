package config

import (
	"log"

	"github.com/hibiken/asynq"
)

var Worker *asynq.Server

func ConnectWorkers() {
	Worker = asynq.NewServer(RedisOpt, asynq.Config{
		Concurrency: 10,
	})

	// Register all task handlers here
	mux := asynq.NewServeMux()

	if err := Worker.Start(mux); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	log.Println("Successfully started and connected workers")
}
