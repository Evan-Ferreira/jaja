package tools

import (
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

type CreateDocxAsyncArgs struct {
	AssignmentName string `json:"assignment_name"`
	AssignmentFileURLs []string `json:"assignment_file_urls"`
	Prompt string `json:"prompt"`
}

func DocxTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name: "docx",
		Description: "Use this tool to create, read, edit, and manipulate Word documents (.docx files).",
	}, CreateDocxAsync)
}


func CreateDocxAsync(ctx tool.Context, args CreateDocxAsyncArgs) (any, error) {
	assignmentId := "ASSIGNMENT-ABC-123"

	return struct{ File string }{ File: "Hello, world!" }, nil
}
