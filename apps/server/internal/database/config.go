package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DBClient *gorm.DB

func ConnectDB() {
	var err error
	dsn := os.Getenv("DB_URL")
	DBClient, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to database")
	}

	log.Println("Successfully connected to database")
}