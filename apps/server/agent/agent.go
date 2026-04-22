package agent

import (
	"context"
	"log"
	"strings"

	"server/agent/agents"

	adkagent "google.golang.org/adk/agent"
	adkrunner "google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

var r *adkrunner.Runner

func ConnectAgent() {
	a, err := agents.CreateOrchestratorAgent(context.Background())
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	r, err = adkrunner.New(adkrunner.Config{
		AppName:           "jaja",
		Agent:             *a,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	log.Println("Agent runner connected")
}

func Run(ctx context.Context, sessionID, userID, prompt string) (string, error) {
	log.Printf("[agent] run started session=%s user=%s", sessionID, userID)
	var result string
	for event, err := range r.Run(ctx, userID, sessionID, genai.NewContentFromText(prompt, genai.RoleUser), adkagent.RunConfig{StreamingMode: adkagent.StreamingModeNone}) {
		if err != nil {
			log.Printf("[agent] error: %v", err)
			return "", err
		}
		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.FunctionCall != nil {
					log.Printf("[agent] %s → call %s", event.Author, part.FunctionCall.Name)
				}
				if part.FunctionResponse != nil {
					log.Printf("[agent] %s → response %s", event.Author, part.FunctionResponse.Name)
				}
			}
		}
		if !event.IsFinalResponse() || event.Content == nil {
			continue
		}
		log.Printf("[agent] final response from %s", event.Author)
		var sb strings.Builder
		for _, part := range event.Content.Parts {
			sb.WriteString(part.Text)
		}
		result = sb.String()
	}
	log.Printf("[agent] run complete session=%s", sessionID)
	return result, nil
}
