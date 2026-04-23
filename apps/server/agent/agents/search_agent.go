package agents

import (
	"context"
	"fmt"

	"server/agent/models"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/geminitool"
)

func NewSearchAgent(ctx context.Context) (agent.Agent, error) {
	model, err := models.NewGeminiModel(ctx, "gemini-2.5-flash")
	if err != nil {
		return nil, fmt.Errorf("create gemini model: %w", err)
	}

	return llmagent.New(llmagent.Config{
		Name:        "search_agent",
		Model:       model,
		Description: "Searches the web for up-to-date information about academic topics, citation formats, subject matter, or any context needed to complete an assignment.",
		Instruction: `You are a research assistant. When asked to search, use the google_search tool to find relevant, accurate, and up-to-date information. Summarize findings clearly.`,
		Tools:       []tool.Tool{geminitool.GoogleSearch{}},
		OutputKey:   "search_results",
	})
}
