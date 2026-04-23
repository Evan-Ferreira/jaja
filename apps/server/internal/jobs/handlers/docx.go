package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"server/internal/database"
	"server/internal/storage"
	"server/internal/util"

	anthropicModels "server/agent/models"
	internalModels "server/internal/models"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/hibiken/asynq"
	"gorm.io/gorm/clause"
)

type DocxPayload struct {
	AssignmentName     string   `json:"assignment_name"`
	AssignmentFileURLs []string `json:"assignment_file_urls"`
	Prompt             string   `json:"prompt"`
	FunctionCallID     string   `json:"function_call_id"`
	SessionID          string   `json:"session_id"`
	UserID             string   `json:"user_id"`
}

func HandleDocx(ctx context.Context, task *asynq.Task) error {
	taskID := task.ResultWriter().TaskID()

	var payload DocxPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("error unmarshalling payload: %w", err)
	}

	queueName := task.Headers()["queue"]
	if queueName == "" {
		queueName = "default"
	}

	job := internalModels.Job{
		ID:      taskID,
		Queue:   internalModels.Queue(queueName),
		Type:    task.Type(),
		Payload: task.Payload(),
		State:   "active",
	}

	if err := database.DBClient.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"state", "updated_at"}),
	}).Create(&job).Error; err != nil {
		util.FailJob(taskID, err)
		return fmt.Errorf("error upserting job state: %w", err)
	}

	claudeService, err := anthropicModels.New(anthropic.ModelClaudeSonnet4_6)
	if err != nil {
		util.FailJob(taskID, err)
		return fmt.Errorf("error creating Claude service: %w", err)
	}

	docs := make([]anthropicModels.PresignedDocument, 0, len(payload.AssignmentFileURLs))
	for _, url := range payload.AssignmentFileURLs {
		docs = append(docs, anthropicModels.PresignedDocument{
			URL:  url,
			Type: anthropicModels.InferDocumentType(url),
		})
	}

	response, err := claudeService.Run(ctx, anthropicModels.AnthropicServiceConfig{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 20000,
		Messages: []anthropicModels.AnthropicMessage{
			{
				Role:    anthropicModels.AnthropicRoleUser,
				Message: payload.Prompt,
			},
		},
		Skills: &[]anthropic.BetaSkillParams{
			{
				SkillID: "docx",
				Type:    anthropic.BetaSkillParamsTypeAnthropic,
				Version: anthropic.String("latest"),
			},
		},
		Documents: docs,
	})
	if respBytes, logErr := json.MarshalIndent(response, "", "  "); logErr != nil {
		fmt.Printf("[docx] claude run response task_id=%s marshal_err=%v dump=%+v\n", taskID, logErr, response)
	} else {
		fmt.Printf("[docx] claude run response task_id=%s\n%s\n", taskID, string(respBytes))
	}
	if err != nil {
		fmt.Printf("[docx] claude run error task_id=%s: %v\n", taskID, err)
		util.FailJob(taskID, err)
		return fmt.Errorf("error running Claude service: %w", err)
	}

	files, err := claudeService.GetFilesFromResponse(ctx, response)
	if err != nil {
		fmt.Printf("[docx] get files from response error task_id=%s: %v\n", taskID, err)
		util.FailJob(taskID, err)
		return fmt.Errorf("error getting files from response: %w", err)
	}

	fmt.Printf("[docx] got %d files from response task_id=%s\n", len(files), taskID)

	filesToUpload := make([]storage.FileToUpload, len(files))
	for i, f := range files {
		filesToUpload[i] = storage.FileToUpload{Filename: f.Filename, MimeType: f.MimeType, Data: f.Data}
	}
	keyPrefix := fmt.Sprintf("agent_outputs/%s/%s", payload.UserID, taskID)
	savedFiles, err := storage.S3BasicsBucket.SaveFilesToS3(ctx, "test-bucket", keyPrefix, filesToUpload)
	if err != nil {
		fmt.Printf("[docx] save files to s3 error task_id=%s: %v\n", taskID, err)
		util.FailJob(taskID, err)
		return fmt.Errorf("error saving files to s3: %w", err)
	}
	fmt.Printf("[docx] saved %d files to s3 task_id=%s\n", len(savedFiles), taskID)

	fmt.Printf("[docx] claude run ok task_id=%s\n", taskID)

	res, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("[docx] marshal response error task_id=%s: %v\n", taskID, err)
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	fmt.Printf("[docx] result bytes task_id=%s len=%d\n", taskID, len(res))
	if _, err = task.ResultWriter().Write(res); err != nil {
		fmt.Printf("[docx] write result error task_id=%s: %v\n", taskID, err)
		return fmt.Errorf("failed to write task result: %w", err)
	}

	database.DBClient.Model(&internalModels.Job{}).Where("id = ?", taskID).
		Updates(map[string]any{"state": "completed", "result": res})

	fmt.Printf("[docx] completed task_id=%s\n", taskID)
	return nil
}

