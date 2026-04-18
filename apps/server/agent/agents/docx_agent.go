package agents

import (
	"fmt"

	"server/agent/models"
	"server/agent/tools"

	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/tool"
)

func createDocxAgent() (llmagent.Config, error) {
	model, err := models.NewAnthropicModel("claude-sonnet-4-6")
	if err != nil {
		return llmagent.Config{}, fmt.Errorf("create model: %w", err)
	}

	docxTool, err := tools.DocxTool()
	if err != nil {
		return llmagent.Config{}, fmt.Errorf("create docx tool: %w", err)
	}

	return llmagent.Config{
		Name:        "docx_agent",
		Model:       model,
		Description: "Writes and generates Word documents (.docx) for academic assignments. Use this agent when the task requires producing a submission-ready Word document.",
		Instruction: `You are JAJA (Just Automate Junk Assignments), an expert academic assistant specializing in completing school assignments to a high standard.
            You will be given assignment files (PDFs, images) and instructions describing what needs to be completed. Your job is to:
            1. Carefully read and analyze all provided assignment materials and instructions.
            2. Produce a thorough, well-structured response that fully satisfies the assignment requirements.
            3. Match the academic level, tone, and formatting conventions expected for the course.
            4. Use the docx tool to generate the final submission as a Word document.

            Always complete the full assignment — do not summarize or skip sections. If the assignment requires specific formatting (headings, citations, word count), follow it exactly.`,
		Tools: []tool.Tool{docxTool},
	}, nil
}
