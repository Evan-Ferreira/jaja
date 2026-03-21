package main

import (
	"fmt"
	"log"
	"os"

	"server/internal/database"
	"server/internal/routes"
	"server/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	database.ConnectDB()
	storage.ConnectObjectStorage()
	//TODO: Not ideal to run seeds on every startup, but this is a temporary measure until we have a better solution for managing test data.

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	log.Println("Successfully loaded environment variables")

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{os.Getenv("FRONTEND_URL")}
	router.Use(cors.New(config))

	api := router.Group("/api")
	{
		routes.RegisterD2LRoutes(api)
	}

	port := os.Getenv("PORT")
	fmt.Println("Starting server on port " + port + "...")

	router.Run(":" + port)
}
