package config

import (
	"log"
	"os"

	"server/internal/tasks"

	"github.com/hibiken/asynq"
)

var (
	Worker    *asynq.Server
	Scheduler *asynq.Scheduler
)

func ConnectWorkers() {
	// Load the Anthropic API key for task handlers that need it.
	AnthropicAPIKey = os.Getenv("ANTHROPIC_API_KEY")

	Worker = asynq.NewServer(RedisOpt, asynq.Config{
		Concurrency: 10,
	})

	// Register all task handlers here
	mux := asynq.NewServeMux()
	checkStatusHandler := tasks.NewCheckAssignmentStatusHandler(DBClient, AnthropicAPIKey)
	mux.HandleFunc(tasks.TypeCheckAssignmentStatus, checkStatusHandler.ProcessTask)

	if err := Worker.Start(mux); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	// Set up periodic task scheduler
	Scheduler = asynq.NewScheduler(RedisOpt, nil)

	task, err := tasks.NewCheckAssignmentStatusTask()
	if err != nil {
		log.Fatalf("Failed to create check assignment status task: %v", err)
	}

	// Run every 2 minutes to check for completed Claude AI batch jobs
	if _, err := Scheduler.Register("*/2 * * * *", task); err != nil {
		log.Fatalf("Failed to register periodic task: %v", err)
	}

	if err := Scheduler.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	log.Println("Successfully started and connected workers")
}
