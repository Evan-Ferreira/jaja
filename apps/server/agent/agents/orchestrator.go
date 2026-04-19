package agents

import (
	"context"
	"fmt"

	"server/agent/models"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/agenttool"
)

func CreateOrchestratorAgent(ctx context.Context) (*agent.Agent, error) {
	model, err := models.NewAnthropicModel("claude-sonnet-4-6")
	if err != nil {
		return nil, fmt.Errorf("create model: %w", err)
	}

	analysisAgent, err := NewAnalysisAgent(ctx)
	if err != nil {
		return nil, fmt.Errorf("create analysis agent: %w", err)
	}

	docxAgent, err := NewDocxAgent()
	if err != nil {
		return nil, fmt.Errorf("create docx agent: %w", err)
	}

	orchestrator, err := llmagent.New(llmagent.Config{
		Name:        "jaja_orchestrator",
		Model:       model,
		Description: "Orchestrates academic assignment completion by routing tasks to specialist agents.",
		Instruction: `You are the JAJA orchestrator. Complete assignments in two steps:
1. Call analysis_agent with the assignment name and file URLs to extract all requirements, rubric, and formatting details.
2. If analysis_agent returns a result starting with "ANALYSIS_FAILED:", immediately stop and return that failure message verbatim. Do NOT call docx_agent or generate any content.
3. Only if analysis succeeded: call docx_agent with the assignment name, file URLs, and a detailed prompt built from the analysis results.

Always run analysis first — never call docx_agent without a successful analysis.`,
		Tools: []tool.Tool{
			agenttool.New(analysisAgent, nil),
			agenttool.New(docxAgent, nil),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("build orchestrator: %w", err)
	}

	return &orchestrator, nil
}
