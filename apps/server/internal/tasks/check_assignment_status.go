package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"server/internal/models"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// TypeCheckAssignmentStatus is the task type for periodic assignment status checks.
const TypeCheckAssignmentStatus = "assignment:check_status"

// CheckAssignmentStatusPayload is the payload for the check assignment status task.
// Empty for periodic tasks since they scan all processing jobs.
type CheckAssignmentStatusPayload struct{}

// NewCheckAssignmentStatusTask creates a new task for checking assignment statuses.
func NewCheckAssignmentStatusTask() (*asynq.Task, error) {
	payload, err := json.Marshal(CheckAssignmentStatusPayload{})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task payload: %w", err)
	}
	return asynq.NewTask(TypeCheckAssignmentStatus, payload), nil
}

// CheckAssignmentStatusHandler holds dependencies for the check assignment status task.
type CheckAssignmentStatusHandler struct {
	DB           *gorm.DB
	AnthropicKey string
}

// NewCheckAssignmentStatusHandler creates a new handler with the required dependencies.
func NewCheckAssignmentStatusHandler(db *gorm.DB, anthropicKey string) *CheckAssignmentStatusHandler {
	return &CheckAssignmentStatusHandler{
		DB:           db,
		AnthropicKey: anthropicKey,
	}
}

// ProcessTask handles the check assignment status task.
// It queries the database for jobs in "processing" state, checks their status
// via the Anthropic Message Batches API, and updates the database accordingly.
func (h *CheckAssignmentStatusHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	log.Println("task: check_assignment_status: starting periodic status check")

	var jobs []models.AssignmentJob
	result := h.DB.
		Where("status = ?", models.AssignmentJobStatusProcessing).
		Where("anthropic_batch_id IS NOT NULL").
		Find(&jobs)

	if result.Error != nil {
		return fmt.Errorf("failed to query processing jobs: %w", result.Error)
	}

	if len(jobs) == 0 {
		log.Println("task: check_assignment_status: no processing jobs found")
		return nil
	}

	log.Printf("task: check_assignment_status: found %d processing jobs to check", len(jobs))

	client, err := h.newAnthropicClient()
	if err != nil {
		return fmt.Errorf("failed to create Anthropic client: %w", err)
	}

	for _, job := range jobs {
		if err := h.checkAndUpdateJob(ctx, client, job); err != nil {
			log.Printf("task: check_assignment_status: error checking job %s: %v", job.ID, err)
			continue
		}
	}

	log.Println("task: check_assignment_status: completed periodic status check")
	return nil
}

// newAnthropicClient creates a new Anthropic API client.
func (h *CheckAssignmentStatusHandler) newAnthropicClient() (*anthropic.Client, error) {
	if h.AnthropicKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is not configured")
	}

	client := anthropic.NewClient(option.WithAPIKey(h.AnthropicKey))
	return &client, nil
}

// checkAndUpdateJob checks the status of a single assignment job via the
// Anthropic Message Batches API and updates the database record.
func (h *CheckAssignmentStatusHandler) checkAndUpdateJob(ctx context.Context, client *anthropic.Client, job models.AssignmentJob) error {
	if job.AnthropicBatchID == nil {
		return fmt.Errorf("job %s has no anthropic_batch_id", job.ID)
	}

	batch, err := client.Messages.Batches.Get(ctx, *job.AnthropicBatchID)
	if err != nil {
		errMsg := fmt.Sprintf("Anthropic API error: %s", err.Error())
		h.DB.Model(&job).Updates(map[string]interface{}{
			"status":        models.AssignmentJobStatusFailed,
			"error_message": errMsg,
			"updated_at":    time.Now(),
		})
		return fmt.Errorf("failed to get batch %s: %w", *job.AnthropicBatchID, err)
	}

	switch batch.ProcessingStatus {
	case anthropic.MessageBatchProcessingStatusEnded:
		return h.handleBatchEnded(job, batch)

	case anthropic.MessageBatchProcessingStatusCanceling:
		log.Printf("task: check_assignment_status: job %s (batch %s) is canceling", job.ID, *job.AnthropicBatchID)

	case anthropic.MessageBatchProcessingStatusInProgress:
		log.Printf("task: check_assignment_status: job %s (batch %s) still in progress (%d processing, %d succeeded)",
			job.ID, *job.AnthropicBatchID, batch.RequestCounts.Processing, batch.RequestCounts.Succeeded)

	default:
		log.Printf("task: check_assignment_status: job %s has unknown batch status: %s", job.ID, batch.ProcessingStatus)
	}

	return nil
}

// handleBatchEnded processes a batch that has finished processing and updates
// the assignment job based on the batch request counts.
func (h *CheckAssignmentStatusHandler) handleBatchEnded(job models.AssignmentJob, batch *anthropic.MessageBatch) error {
	counts := batch.RequestCounts

	if counts.Succeeded > 0 && counts.Errored == 0 && counts.Expired == 0 && counts.Canceled == 0 {
		// All requests succeeded
		resultSummary := fmt.Sprintf(
			"Batch completed successfully. %d requests succeeded.",
			counts.Succeeded,
		)
		h.DB.Model(&job).Updates(map[string]interface{}{
			"status":     models.AssignmentJobStatusCompleted,
			"result":     resultSummary,
			"updated_at": time.Now(),
		})
		log.Printf("task: check_assignment_status: job %s completed successfully", job.ID)

	} else if counts.Canceled > 0 && counts.Succeeded == 0 && counts.Errored == 0 {
		// All cancelled
		h.DB.Model(&job).Updates(map[string]interface{}{
			"status":        models.AssignmentJobStatusCancelled,
			"error_message": "Batch was cancelled",
			"updated_at":    time.Now(),
		})
		log.Printf("task: check_assignment_status: job %s was cancelled", job.ID)

	} else if counts.Expired > 0 && counts.Succeeded == 0 && counts.Errored == 0 {
		// All expired
		h.DB.Model(&job).Updates(map[string]interface{}{
			"status":        models.AssignmentJobStatusExpired,
			"error_message": "Batch expired before completion",
			"updated_at":    time.Now(),
		})
		log.Printf("task: check_assignment_status: job %s expired", job.ID)

	} else if counts.Errored > 0 {
		// Some or all errored
		errMsg := fmt.Sprintf(
			"Batch ended with errors: %d succeeded, %d errored, %d expired, %d canceled",
			counts.Succeeded, counts.Errored, counts.Expired, counts.Canceled,
		)
		h.DB.Model(&job).Updates(map[string]interface{}{
			"status":        models.AssignmentJobStatusFailed,
			"error_message": errMsg,
			"updated_at":    time.Now(),
		})
		log.Printf("task: check_assignment_status: job %s failed: %s", job.ID, errMsg)

	} else {
		// Mixed results
		resultSummary := fmt.Sprintf(
			"Batch ended: %d succeeded, %d errored, %d expired, %d canceled",
			counts.Succeeded, counts.Errored, counts.Expired, counts.Canceled,
		)
		h.DB.Model(&job).Updates(map[string]interface{}{
			"status":     models.AssignmentJobStatusCompleted,
			"result":     resultSummary,
			"updated_at": time.Now(),
		})
		log.Printf("task: check_assignment_status: job %s completed with mixed results: %s", job.ID, resultSummary)
	}

	return nil
}
