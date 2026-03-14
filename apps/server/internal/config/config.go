package config

import (
	"log"

	"github.com/joho/godotenv"
)

// TODO: This will need to change when we consider different d2l orgs
const D2LBaseURL = "https://onq.queensu.ca/"

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	log.Println("Successfully loaded environment variables")
}
