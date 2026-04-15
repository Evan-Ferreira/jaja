package runner

import (
	"context"
	"fmt"
	"strings"

	mainAgent "server/agent"

	"github.com/google/uuid"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

type Runner struct {
	Agent *agent.Agent
	UserID uuid.UUID
	Runner *runner.Runner
}

// RunInput holds the context passed into a single agent invocation.
type RunInput struct {
	SessionID string
	Prompt    string
}

// TODO: wip agent runner, JAJA for now
// Run wires up a one-shot ADK runner, sends the prompt to the JAJA agent,
// and returns the concatenated final text response.
func (r *Runner) New() error {
	var err error
	r.Agent, err = mainAgent.CreateJAJAAgent()
	if err != nil {
		return fmt.Errorf("create agent: %w", err)
	}

	// TODO: Look into what all these settings do, like sessions, artifacts, memory, subagents, userId, etc.
	r.Runner, err = runner.New(runner.Config{
		AppName:           "jaja",
		Agent:             *r.Agent,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})

	if err != nil {
		return fmt.Errorf("create runner: %w", err)
	}

	return nil
}

func (r *Runner) Run(ctx context.Context, input RunInput) (string, error) {
	msg := genai.NewContentFromText(input.Prompt, genai.RoleUser)

	var sb strings.Builder
	for event, err := range r.Runner.Run(ctx, r.UserID.String(), input.SessionID, msg, agent.RunConfig{}) {
		if err != nil {
			return "", fmt.Errorf("agent run: %w", err)
		}
		if event.LLMResponse.Content == nil {
			continue
		}
		for _, part := range event.LLMResponse.Content.Parts {
			sb.WriteString(part.Text)
		}
	}


	return sb.String(), nil
}