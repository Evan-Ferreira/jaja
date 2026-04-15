package tools

import (
	"encoding/json"
	"fmt"
	"server/internal/jobs"
	"server/internal/util"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

type CreateDocxAsyncArgs struct {
	AssignmentName     string   `json:"assignment_name"`
	AssignmentFileURLs []string `json:"assignment_file_urls"`
	Prompt             string   `json:"prompt"`
}

type docxJobPayload struct {
	CreateDocxAsyncArgs
	FunctionCallID string `json:"function_call_id"`
	UserID         string `json:"user_id"`
	SessionID      string `json:"session_id"`
}

type CreateDocxAsyncResults struct {
	TaskID         string `json:"task_id"`
	FunctionCallID string `json:"function_call_id"`
	State          string `json:"state"`
}

func DocxTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "create_docx",
		Description: "Use this tool to create, read, edit, and manipulate Word documents (.docx files).",
	}, CreateDocxAsync)
}

func CreateDocxAsync(ctx tool.Context, args CreateDocxAsyncArgs) (any, error) {
	jobPayload := docxJobPayload{
		CreateDocxAsyncArgs: args,
		FunctionCallID:      ctx.FunctionCallID(),
		UserID:              ctx.UserID(),
		SessionID:           ctx.SessionID(),
	}

	payload, err := json.Marshal(jobPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments: %w", err)
	}

	taskID, err := util.EnqueueJob(jobs.JobTypeDocx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue job: %w", err)
	}

	return &CreateDocxAsyncResults{TaskID: taskID, FunctionCallID: ctx.FunctionCallID(), State: "pending"}, nil
}
