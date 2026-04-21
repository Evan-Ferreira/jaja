package util

import (
	"context"
	"encoding/json"
	"fmt"
	"server/internal/database"
	"server/internal/models"
	"server/internal/queue"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const DefaultPollTimeout = 10 * time.Minute
const DefaultPollInterval = 5 * time.Second
const DefaultRetention = 15 * time.Minute

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

	taskInfo, err := inspector.GetTaskInfo(string(models.QueueDefault), taskID)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return nil
	}
	return taskInfo
}

func PollJob(ctx context.Context, taskID string, maxTimeout ...time.Duration) (*asynq.TaskInfo, error) {
	timeout := DefaultPollTimeout
	if len(maxTimeout) > 0 {
		timeout = maxTimeout[0]
	}

	fmt.Printf("Starting poll for taskID: %q (timeout: %s, interval: %s)\n", taskID, timeout, DefaultPollInterval)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		taskInfo := FetchJobByID(taskID)
		if taskInfo == nil {
			return nil, fmt.Errorf("fetching job %s: not found", taskID)
		}

		switch taskInfo.State {
		case asynq.TaskStateCompleted:
			fmt.Printf("Job COMPLETED after %d polls — result length: %d\n", taskInfo.ID, len(taskInfo.Result))
			return taskInfo, nil
		case asynq.TaskStateArchived:
			fmt.Printf("Job ARCHIVED/FAILED for taskID: %q — lastErr: %s\n", taskID, taskInfo.LastErr)
			return taskInfo, nil
		}

		select {
		case <-time.After(DefaultPollInterval):
		case <-ctx.Done():
			fmt.Printf("TIMED OUT for taskID: %q (%s)\n", taskID, timeout)
			return nil, fmt.Errorf("polling timed out after %s", timeout)
		}
	}
}

func EnqueueJob(taskType string, payload []byte) (string, error) {
	id := uuid.NewString()
	task := asynq.NewTask(taskType, payload, asynq.TaskID(id))

	dbJob := &models.Job{
		ID:      id,
		Queue:   models.QueueDefault,
		Type:    taskType,
		Payload: json.RawMessage(payload),
		State:   models.JobStatePending,
	}

	fmt.Printf("Payload type: %T\n", json.RawMessage(payload))

	if err := database.DBClient.Create(dbJob).Error; err != nil {
		return "", fmt.Errorf("persist job %s: %w", taskType, err)
	}

	job, err := queue.RedisClient.Enqueue(task, asynq.Retention(DefaultRetention))
	if err != nil {
		database.DBClient.Delete(&models.Job{}, "id = ?", id)
		fmt.Printf("ERROR: %v\n", err)
		return "", fmt.Errorf("enqueue %s: %w", taskType, err)
	}

	fmt.Printf("Enqueued successfully — jobID: %q\n", job.ID)
	return job.ID, nil
}

func FailJob(taskID string, err error) *models.Job {
	database.DBClient.Model(&models.Job{}).Where("id = ?", taskID).
		Updates(map[string]any{"state": models.JobStateArchived, "last_err": err.Error()})
	return FetchJobFromDB(taskID)
}

func CompleteJob(taskID string, result []byte) *models.Job {
	database.DBClient.Model(&models.Job{}).Where("id = ?", taskID).
		Updates(map[string]any{"state": models.JobStateCompleted, "result": result})
	return FetchJobFromDB(taskID)
}