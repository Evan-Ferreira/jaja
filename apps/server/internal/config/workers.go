package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"server/internal/tasks"

	"github.com/hibiken/asynq"
)

var Worker *asynq.Server

func ConnectWorkers() {
	Worker = asynq.NewServer(RedisOpt, asynq.Config{
		Concurrency: 10,
	})

	// Register all task handlers here
	mux := asynq.NewServeMux()
	mux.HandleFunc("agent:assignment", handler)
	
	if err := Worker.Start(mux); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	log.Println("Successfully started and connected workers")
}

func handler(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
		case "agent:assignment":
			var p tasks.AssignmentTaskPayload
			if err := json.Unmarshal(task.Payload(), &p); err != nil {
				return err
			}
			log.Printf(" [*] Complete Assignment %s", p.AssignmentID)
		default:
			return fmt.Errorf("unexpected task type: %s", task.Type())
		}
		return nil
	}
