package runner

import (
	"context"
	"fmt"

	orchestrator "server/agent/agents"
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

func (r *AgentRunner) runTurn(ctx context.Context, sessionId string, userId string, content *genai.Content) string {
    funcCallID := ""

    for event, err := range r.Runner.Run(ctx, userId, sessionId, content, agent.RunConfig{
        StreamingMode: agent.StreamingModeNone,
    }) {
        if err != nil {
            continue
        }
        for _, part := range event.Content.Parts {
            if fc := part.FunctionCall; fc != nil {
                if fc.Name == "create_docx" {
                    funcCallID = fc.ID
                }
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
	
	job, err := util.GetJobByFunctionCallID(funcCallID)
	if err != nil {
		return "", fmt.Errorf("ERROR: Failed to get job by function call id: %v", err)
	}

	jobInfo, err := util.PollJob(ctx, job.ID)
	if err != nil {
		return "", fmt.Errorf("ERROR: Failed to poll job: %v", err)
	}

	willContinue := false
	docxStatusResponse := &genai.FunctionResponse{
		Name: "create_docx",
		ID:   funcCallID,
		Response: map[string]any{
			"task_id":          job.ID,
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
