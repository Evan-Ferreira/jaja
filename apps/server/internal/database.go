package internal

import (
	"log"
	"os"

	"server/internal/models"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB(){
	var err error
	dsn := os.Getenv("DB_URL")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to database")
	}

	log.Println("Successfully connected to database")

	seedDB()
}

func seedDB() {
	testUser := models.User{
		ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
	}

	result := DB.FirstOrCreate(&testUser)
	if result.Error != nil {
		log.Printf("Failed to seed test user (migrations may not have run yet): %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Println("Test user created successfully")
	} else {
		log.Println("Test user already exists, skipping")
	}
}