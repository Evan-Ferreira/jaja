package tools

import (
	"fmt"

	"server/internal/services"
	"server/internal/storage"

	"github.com/anthropics/anthropic-sdk-go"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

type CreateDocxArgs struct {
	AssignmentName     string   `json:"assignment_name"`
	AssignmentFileURLs []string `json:"assignment_file_urls"`
	Prompt             string   `json:"prompt"`
}

type CreateDocxResult struct {
	Files  []storage.SavedFileResult `json:"files"`
	Status string                    `json:"status"`
}

func DocxTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "create_docx",
		Description: "Generate a Word document (.docx) for an academic assignment. Returns download URLs for the generated files.",
	}, CreateDocx)
}

func CreateDocx(ctx tool.Context, args CreateDocxArgs) (*CreateDocxResult, error) {
	claudeService, err := services.New()
	if err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	docs := make([]services.PresignedDocument, len(args.AssignmentFileURLs))
	for i, url := range args.AssignmentFileURLs {
		docs[i] = services.PresignedDocument{URL: url, Type: services.InferDocumentType(url)}
	}

	response, err := claudeService.Run(ctx, services.ClaudeServiceConfig{
		Model: "claude-sonnet-4-6",
		//TODO: make this adjustable
		MaxTokens: 20000,
		Messages: []services.AnthropicMessage{{
			Role:    services.AnthropicRoleUser,
			Message: args.Prompt,
		}},
		Skills: &[]anthropic.BetaSkillParams{{
			SkillID: "docx",
			Type:    anthropic.BetaSkillParamsTypeAnthropic,
			Version: anthropic.String("latest"),
		}},
		Documents: docs,
	})
	if err != nil {
		return nil, fmt.Errorf("run claude: %w", err)
	}

	files, err := claudeService.GetFilesFromResponse(ctx, response)
	if err != nil {
		return nil, fmt.Errorf("get files: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files generated")
	}

	toUpload := make([]storage.FileToUpload, len(files))
	for i, f := range files {
		toUpload[i] = storage.FileToUpload{Filename: f.Filename, MimeType: f.MimeType, Data: f.Data}
	}

	keyPrefix := fmt.Sprintf("agent_outputs/%s/%s", ctx.UserID(), ctx.FunctionCallID())
	saved, err := storage.S3BasicsBucket.SaveFilesToS3(ctx, "test-bucket", keyPrefix, toUpload)
	if err != nil {
		return nil, fmt.Errorf("save to s3: %w", err)
	}

	return &CreateDocxResult{Files: saved, Status: "completed"}, nil
}
