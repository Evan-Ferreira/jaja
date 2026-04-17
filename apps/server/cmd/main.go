package main

import (
	"fmt"
	"log"
	"os"
	"server/internal/database"
	"server/internal/queue"
	"server/internal/routes"
	"server/internal/storage"
	"server/internal/workers"

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

	queue.ConnectRedis()
	database.ConnectDB()
	storage.ConnectObjectStorage()
	workers.Connect()

	defer queue.RedisClient.Close()
	defer workers.Server.Shutdown()

	router := gin.Default()

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		log.Println("WARNING: FRONTEND_URL is not set — CORS will reject all cross-origin requests")
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{frontendURL}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	router.Use(cors.New(corsConfig))

	api := router.Group("/")
	{
		routes.RegisterD2LRoutes(api)

		// Only register dev/debug routes in non-release mode.
		if gin.Mode() != gin.ReleaseMode {
			routes.RegisterDevRoutes(api)
		} else {
			log.Println("Running in release mode — dev routes are disabled")
		}
	}

	port := os.Getenv("PORT")
	fmt.Println("Starting server on port " + port + "...")

	router.Run(":" + port)

}
