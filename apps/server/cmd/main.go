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

	config.ConnectRedis()
	config.ConnectDB()
	config.ConnectObjectStorage()
	config.ConnectWorkers()

	defer config.RedisClient.Close()
	defer config.Worker.Shutdown()

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
