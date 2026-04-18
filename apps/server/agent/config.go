package agent

import (
	"log"
	"server/agent/runner"
)

var AgentRunner *runner.AgentRunner
var OrchestratedRunner *runner.AgentRunner

func ConnectAgent() {
	r, err := runner.Init()
	if err != nil {
		log.Fatalf("Failed to create agent runner: %v", err)
	}
	log.Println("Successfully created agent runner")
	AgentRunner = r
}

func ConnectOrchestratedAgent() {
	r, err := runner.InitOrchestrated()
	if err != nil {
		log.Fatalf("Failed to create orchestrated agent runner: %v", err)
	}
	log.Println("Successfully created orchestrated agent runner")
	OrchestratedRunner = r
}

