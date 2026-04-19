package agents

import (
	"context"
	"fmt"

	"server/agent/models"
	"server/agent/tools"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/tool"
)

func NewAnalysisAgent(ctx context.Context) (agent.Agent, error) {
	model, err := models.NewAnthropicModel("claude-haiku-4-5-20251001")
	if err != nil {
		return nil, fmt.Errorf("create model: %w", err)
	}

	analyzeTool, err := tools.AnalyzeTool()
	if err != nil {
		return nil, fmt.Errorf("create analyze tool: %w", err)
	}

	return llmagent.New(llmagent.Config{
		Name:        "analysis_agent",
		Model:       model,
		Description: "Analyzes assignment documents to extract requirements, rubric, formatting rules, and all relevant details needed for completion.",
		Instruction: `You are a document analysis specialist. Your ONLY job is to extract details from provided documents.

STRICT RULES — follow exactly:
1. Call analyze_assignment with the provided assignment name and file URLs.
2. If analyze_assignment returns ANY error, failure, or empty result — immediately output EXACTLY: "ANALYSIS_FAILED: <error message>" and STOP. Do NOT call any other tools. Do NOT generate any content.
3. If analyze_assignment succeeds, return a structured summary covering: assignment type, requirements, topics, length, formatting, rubric, and any other constraints.

NEVER generate, write, or draft any assignment content. NEVER attempt to complete the assignment. If the documents cannot be read for any reason, FAIL immediately with "ANALYSIS_FAILED:" and nothing else.`,
		Tools:     []tool.Tool{analyzeTool},
		OutputKey: "assignment_analysis",
	})
}
