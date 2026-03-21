package main

import (
	"fmt"
	"log"
	"os"

	"server/internal/config"
	"server/internal/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	log.Println("Successfully loaded environment variables")

	config.ConnectDB()
	config.ConnectObjectStorage()

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
