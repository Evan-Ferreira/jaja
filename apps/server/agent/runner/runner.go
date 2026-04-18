package runner

import (
	"context"
	"fmt"

	orchestrator "server/agent/agents"
	"server/agent/tools"
	"server/internal/util"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

type AgentRunner struct {
	Runner *runner.Runner
}

// RunInput holds the context passed into a single agent invocation.
type RunInput struct {
	SessionID string
	Prompt    string
	UserID    string
}

// TODO: wip agent runner, JAJA for now
// Run wires up a one-shot ADK runner, sends the prompt to the JAJA agent,
// and returns the concatenated final text response.
func Init() (*AgentRunner, error) {
	var err error

	agent, err := orchestrator.CreateJAJAAgent()
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	runner, err := runner.New(runner.Config{
		AppName:           "jaja",
		Agent:             *agent,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})

	if err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}

	return &AgentRunner{
		Runner: runner,
	}, nil
}

func InitOrchestrated() (*AgentRunner, error) {
	a, err := orchestrator.CreateOrchestratorAgent()
	if err != nil {
		return nil, fmt.Errorf("create orchestrator agent: %w", err)
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

func (r *AgentRunner) RunOrchestrated(ctx context.Context, input RunInput) (string, error) {
	fmt.Println("[RunOrchestrated] === Turn 1: Sending user prompt to orchestrator ===")
	funcCallID := r.runTurn(ctx, input.UserID, input.SessionID, genai.NewContentFromText(input.Prompt, genai.RoleUser))
	if funcCallID == "" {
		return "", fmt.Errorf("ERROR: Tool 'create_docx' not called in Turn 1.")
	}
	jobID, ok := tools.ClaimJobID(funcCallID)
	if !ok {
		return "", fmt.Errorf("ERROR: No job ID registered for funcCallID %s", funcCallID)
	}
	fmt.Printf("[RunOrchestrated] funcCallID: %s, jobID: %s\n", funcCallID, jobID)

	jobInfo, err := util.PollJob(ctx, jobID)
	if err != nil {
		return "", fmt.Errorf("ERROR: Failed to poll job: %v", err)
	}

	willContinue := false
	docxStatusResponse := &genai.FunctionResponse{
		Name: "create_docx",
		ID:   funcCallID,
		Response: map[string]any{
			"task_id":          jobID,
			"function_call_id": funcCallID,
			"state":            "completed",
			"result":           jobInfo.Result,
		},
		WillContinue: &willContinue,
	}

	fmt.Println("[RunOrchestrated] === Turn 2: Sending function response back to agent ===")
	appResponseWithStatus := &genai.Content{
		Role:  string(genai.RoleUser),
		Parts: []*genai.Part{{FunctionResponse: docxStatusResponse}},
	}
	r.runTurn(ctx, input.UserID, input.SessionID, appResponseWithStatus)
	fmt.Printf("[RunOrchestrated] Turn 2 complete — returning result (length: %d)\n", len(jobInfo.Result))
	fmt.Println("---------- [RunOrchestrated] Agent run complete ----------\n")
	return string(jobInfo.Result), nil
}

func (r *AgentRunner) runTurn(ctx context.Context, sessionId, userId string, content *genai.Content) string {
	var funcCallID string

	for event, err := range r.Runner.Run(ctx, userId, sessionId, content, agent.RunConfig{StreamingMode: agent.StreamingModeNone}) {
		if err != nil {
			continue
		}
		fmt.Printf("[runTurn] event from agent: %q\n", event.Author)
		for _, part := range event.Content.Parts {
			if fc := part.FunctionCall; fc != nil && fc.Name == "create_docx" {
				funcCallID = fc.ID
			}
		}
	}
	return funcCallID
}

func (r *AgentRunner) Run(ctx context.Context, input RunInput) (string, error) {
	fmt.Println("[Run] === Turn 1: Sending user prompt to agent ===")
	funcCallID := r.runTurn(ctx, input.UserID, input.SessionID, genai.NewContentFromText(input.Prompt, genai.RoleUser))
	if funcCallID == "" {
		return "", fmt.Errorf("ERROR: Tool 'create_docx' not called in Turn 1.")
	}
	jobID, ok := tools.ClaimJobID(funcCallID)
	if !ok {
		return "", fmt.Errorf("ERROR: No job ID registered for funcCallID %s", funcCallID)
	}
	fmt.Printf("[Run] funcCallID: %s, jobID: %s\n", funcCallID, jobID)

	jobInfo, err := util.PollJob(ctx, jobID)
	if err != nil {
		return "", fmt.Errorf("ERROR: Failed to poll job: %v", err)
	}

	willContinue := false
	docxStatusResponse := &genai.FunctionResponse{
		Name: "create_docx",
		ID:   funcCallID,
		Response: map[string]any{
			"task_id":          jobID,
			"function_call_id": funcCallID,
			"state":            "completed",
			"result":           jobInfo.Result,
		},
		WillContinue: &willContinue,
	}

	fmt.Println("[Run] === Turn 2: Sending function response back to agent ===")
	appResponseWithStatus := &genai.Content{
		Role:  string(genai.RoleUser),
		Parts: []*genai.Part{{FunctionResponse: docxStatusResponse}},
	}
	r.runTurn(ctx, input.UserID, input.SessionID, appResponseWithStatus)
	fmt.Printf("[Run] Turn 2 complete — returning result (length: %d)\n", len(jobInfo.Result))
	fmt.Println("---------- [Run] Agent run complete ----------\n")
	return string(jobInfo.Result), nil
}
