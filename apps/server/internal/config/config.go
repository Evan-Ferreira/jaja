package config

import (
	"log"

	"github.com/joho/godotenv"
)

// AnthropicAPIKey holds the Anthropic API key loaded from the environment.
// Set during ConnectWorkers() after env vars are loaded.
var AnthropicAPIKey string

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	log.Println("Successfully loaded environment variables")
}
