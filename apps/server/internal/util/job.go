package util

import (
	"fmt"
	"server/internal/config"
	"server/internal/jobs"
	"server/internal/models"

	"github.com/google/uuid"
)

func FetchJobByID(jobID uuid.UUID) *models.Job {
	var job models.Job
	if err := config.DBClient.First(&job, jobID).Error; err != nil {
		fmt.Println("Error fetching job:", err)
		return nil
	}
	return &job
}

func FailJobById(jobID uuid.UUID, err error) *models.Job {
	config.DBClient.Model(&models.Job{}).Where("id = ?", jobID).
		Updates(map[string]any{"status": jobs.JobStatusFailed, "error": err.Error()})
	return FetchJobByID(jobID)
}