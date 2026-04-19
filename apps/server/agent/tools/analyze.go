package tools

import (
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"server/internal/services"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

type AnalyzeAssignmentArgs struct {
	AssignmentName     string   `json:"assignment_name"`
	AssignmentFileURLs []string `json:"assignment_file_urls"`
}

type AnalysisResult struct {
	Analysis string `json:"analysis"`
}

func AnalyzeTool() (tool.Tool, error) {
	return functiontool.New(functiontool.Config{
		Name:        "analyze_assignment",
		Description: "Download and read assignment documents from the provided URLs to extract requirements, rubric, formatting instructions, and any other relevant details.",
	}, AnalyzeAssignment)
}

func AnalyzeAssignment(ctx tool.Context, args AnalyzeAssignmentArgs) (*AnalysisResult, error) {
	svc, err := services.New()
	if err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	// Documents first, instruction text last — all in one user message.
	blocks := make([]anthropic.BetaContentBlockParamUnion, 0, len(args.AssignmentFileURLs)+1)
	for _, url := range args.AssignmentFileURLs {
		switch services.InferDocumentType(url) {
		case services.DocumentTypePDF:
			blocks = append(blocks, anthropic.NewBetaDocumentBlock(anthropic.BetaURLPDFSourceParam{URL: url}))
		default:
			blocks = append(blocks, anthropic.NewBetaImageBlock(anthropic.BetaURLImageSourceParam{URL: url}))
		}
	}
	blocks = append(blocks, anthropic.NewBetaTextBlock(`Analyze this assignment and extract the following in detail:
1. Assignment type (essay, lab report, case study, etc.)
2. Full requirements and instructions
3. Key topics that must be covered
4. Word count or length requirements
5. Formatting requirements (citation style, headings, font, etc.)
6. Rubric or grading criteria
7. Any other constraints or notes

Be thorough — this analysis will be used to write the full assignment.`))

	response, err := svc.Run(ctx, services.ClaudeServiceConfig{
		Model:     anthropic.ModelClaudeHaiku4_5_20251001,
		MaxTokens: 2048,
		Messages: []services.AnthropicMessage{{
			Role:          services.AnthropicRoleUser,
			ContentBlocks: blocks,
		}},
		Betas: &[]anthropic.AnthropicBeta{anthropic.AnthropicBetaPDFs2024_09_25},
	})
	if err != nil {
		return nil, fmt.Errorf("analyze documents: %w", err)
	}

	var analysis string
	for _, block := range response.Content {
		if text, ok := block.AsAny().(anthropic.BetaTextBlock); ok {
			analysis += text.Text
		}
	}

	if analysis == "" {
		return nil, fmt.Errorf("no analysis returned from model")
	}

	return &AnalysisResult{Analysis: analysis}, nil
}
