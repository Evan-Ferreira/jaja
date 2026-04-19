package runner

import (
	"context"
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/genai"
)

type AgentRunner struct {
	Runner *runner.Runner
}

type RunInput struct {
	SessionID string
	Prompt    string
	UserID    string
}

func (r *AgentRunner) runTurn(ctx context.Context, sessionID, userID string, content *genai.Content) string {
	var text string
	for event, err := range r.Runner.Run(ctx, userID, sessionID, content, agent.RunConfig{StreamingMode: agent.StreamingModeNone}) {
		if err != nil || event.Content == nil {
			continue
		}
		fmt.Printf("[runTurn] agent: %q\n", event.Author)
		for _, part := range event.Content.Parts {
			if part.Text != "" {
				text += part.Text
			}
		}
	}
	return text
}

func (r *AgentRunner) Run(ctx context.Context, input RunInput) (string, error) {
	fmt.Println("[Run] starting")
	return r.runTurn(ctx, input.SessionID, input.UserID, genai.NewContentFromText(input.Prompt, genai.RoleUser)), nil
}
