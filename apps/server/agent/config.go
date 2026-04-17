package agent

import (
	"log"
	"server/agent/runner"
)

var AgentRunner *runner.AgentRunner

func ConnectAgent() {
	r, err := runner.Init() 
	if err != nil { log.Fatalf("Failed to create agent runner: %v", err) }                                                                                                                 
	log.Println("Successfully created agent runner")
	AgentRunner = r                         
}

