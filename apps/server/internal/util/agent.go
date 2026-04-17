package util

import (
	"fmt"
	"server/internal/database"
	"server/internal/models"
)

func GetJobByFunctionCallID(funcCallID string) (*models.Job, error) {
	fmt.Printf("[GetJobByFunctionCallID] Querying DB for funcCallID: %q\n", funcCallID)
	var job models.Job
	if err := database.DBClient.Where("payload->>'function_call_id' = ?", funcCallID).First(&job).Error; err != nil {
		fmt.Printf("[GetJobByFunctionCallID] ERROR: %v\n", err)
		return nil, err
	}
	fmt.Printf("[GetJobByFunctionCallID] Found job — ID: %q, State: %q\n", job.ID, job.State)
	return &job, nil
}
