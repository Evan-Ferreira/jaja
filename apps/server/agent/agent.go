package agent

import (
	"context"
	"log"
	"os"

	"server/agent/models"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/tool"
)

// TODO: Change agent boilerplate
func Run() {
    ctx := context.Background()

	model, err := models.NewAnthropicModel("claude-sonnet-4-20250514")

    if err != nil {
        log.Fatalf("Failed to create model: %v", err)
    }

    // 2. Define the agent.
    a, err := llmagent.New(llmagent.Config{
        Name:        "multi_tool_agent",
        Model:       model,
        Description: "An agent that can answer questions using Google Search.",
        Instruction: "You are a helpful assistant. Use the available tools to answer questions.",
        Tools: []tool.Tool{
        },
    })
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // 3. Configure the launcher and run.
    config := &launcher.Config{
        AgentLoader: agent.NewSingleLoader(a),
    }

    l := full.NewLauncher()
    if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
        log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
    }
}