package util

import (
	"context"
	"fmt"
	"server/internal/database"
	"server/internal/models"
	"server/internal/queue"
	"time"

	"github.com/hibiken/asynq"
)

const DEFAULT_POLL_TIMEOUT = 10 * time.Minute
const POLL_INTERVAL = 5 * time.Second
const DEFAULT_RETENTION = 15 * time.Minute

func FetchJobFromDB(taskID string) *models.Job {
	var job models.Job
	if err := database.DBClient.First(&job, "id = ?", taskID).Error; err != nil {
		return nil
	}
	return &job
}

func FetchJobByID(taskID string) (*asynq.TaskInfo) {
	inspector := asynq.NewInspector(queue.RedisOpt)
	defer inspector.Close()

	taskInfo, err := inspector.GetTaskInfo("default", taskID)
	if err != nil {
		fmt.Printf("[FetchJobByID] ERROR: %v\n", err)
		return nil
	}
	return taskInfo
}

func PollJob(ctx context.Context, taskID string, maxTimeout ...time.Duration) (*asynq.TaskInfo, error) {
	timeout := DEFAULT_POLL_TIMEOUT
	if len(maxTimeout) > 0 {
		timeout = maxTimeout[0]
	}

	fmt.Printf("[PollJob] Starting poll for taskID: %q (timeout: %s, interval: %s)\n", taskID, timeout, POLL_INTERVAL)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		taskInfo := FetchJobByID(taskID)
		if taskInfo == nil {
			return nil, fmt.Errorf("fetching job %s: not found", taskID)
		}

		switch taskInfo.State {
		case asynq.TaskStateCompleted:
			fmt.Printf("[PollJob] Job COMPLETED after %d polls — result length: %d\n", taskInfo.ID, len(taskInfo.Result))
			return taskInfo, nil
		case asynq.TaskStateArchived:
			fmt.Printf("[PollJob] Job ARCHIVED/FAILED for taskID: %q — lastErr: %s\n", taskID, taskInfo.LastErr)
			return taskInfo, nil
		}

		select {
		case <-time.After(POLL_INTERVAL):
		case <-ctx.Done():
			fmt.Printf("[PollJob] TIMED OUT for taskID: %q (%s)\n", taskID, timeout)
			return nil, fmt.Errorf("polling timed out after %s", timeout)
		}
	}
}

func EnqueueJob(taskType string, payload []byte) (string, error) {
	task := asynq.NewTask(taskType, payload)

	job, err := queue.RedisClient.Enqueue(task, asynq.Retention(DEFAULT_RETENTION))
	if err != nil {
		fmt.Printf("[EnqueueJob] ERROR: %v\n", err)
		return "", fmt.Errorf("enqueue %s: %w", taskType, err)
	}

	fmt.Printf("[EnqueueJob] Enqueued successfully — jobID: %q\n", job.ID)
	return job.ID, nil
}

func FailJob(taskID string, err error) *models.Job {
	database.DBClient.Model(&models.Job{}).Where("id = ?", taskID).
		Updates(map[string]any{"state": "archived", "last_err": err.Error()})
	return FetchJobFromDB(taskID)
}
