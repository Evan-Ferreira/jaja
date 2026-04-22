package main

import (
	"context"
	"log"
	"os"

	"server/agent/agents"
	"server/internal/storage"

	"github.com/joho/godotenv"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
)

func main() {
	for _, f := range []string{".env", "../.env"} {
		if err := godotenv.Load(f); err == nil {
			break
		}
	}

	storage.ConnectObjectStorage()

	ctx := context.Background()

	orchestrator, err := agents.CreateOrchestratorAgent(ctx)
	if err != nil {
		log.Fatalf("Failed to create orchestrator agent: %v", err)
	}

	analysisAgent, err := agents.NewAnalysisAgent(ctx)
	if err != nil {
		log.Fatalf("Failed to create analysis agent: %v", err)
	}

	docxAgent, err := agents.NewDocxAgent()
	if err != nil {
		log.Fatalf("Failed to create docx agent: %v", err)
	}

	loader, err := agent.NewMultiLoader(*orchestrator, analysisAgent, docxAgent)
	if err != nil {
		log.Fatalf("Failed to create agent loader: %v", err)
	}

	config := &launcher.Config{
		AgentLoader: loader,
	}

	l := full.NewLauncher()
	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
