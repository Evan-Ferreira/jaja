package runner

import (
	"context"
	"fmt"

	orchestrator "server/agent/agents"

	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
)

func Init() (*AgentRunner, error) {
	a, err := orchestrator.CreateOrchestratorAgent(context.Background())
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	r, err := runner.New(runner.Config{
		AppName:           "jaja",
		Agent:             *a,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}

	return &AgentRunner{Runner: r}, nil
}
