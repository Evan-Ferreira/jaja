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

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{os.Getenv("FRONTEND_URL")}
	router.Use(cors.New(config))

	api := router.Group("/")
	{
		routes.RegisterD2LRoutes(api)
		routes.RegisterDevRoutes(api)
	}

	port := os.Getenv("PORT")
	fmt.Println("Starting server on port " + port + "...")

	router.Run(":" + port)

}
