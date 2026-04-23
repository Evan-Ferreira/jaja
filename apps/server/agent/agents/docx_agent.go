package agents

import (
	"fmt"

	"server/agent/models"
	"server/agent/tools"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/tool"
)

func NewDocxAgent() (agent.Agent, error) {
	model, err := models.NewAnthropicModel("claude-sonnet-4-6")
	if err != nil {
		return nil, fmt.Errorf("create model: %w", err)
	}

	docxTool, err := tools.DocxTool()
	if err != nil {
		return nil, fmt.Errorf("create docx tool: %w", err)
	}

	return llmagent.New(llmagent.Config{
		Name:        "docx_agent",
		Model:       model,
		Description: "Generates a complete Word document (.docx) for an academic assignment given files and instructions.",
		Instruction: `You are JAJA (Just Automate Junk Assignments), an expert academic assistant.
You will receive assignment materials and instructions. Your job is to:
1. Carefully read and analyze all provided materials.
2. Produce a thorough, well-structured response that fully satisfies the assignment requirements.
3. Match the academic level, tone, and formatting conventions expected.
4. Call create_docx to generate the final Word document submission.

Always complete the full assignment — do not summarize or skip sections.`,
		Tools:     []tool.Tool{docxTool},
		OutputKey: "docx_output",
	})
}
