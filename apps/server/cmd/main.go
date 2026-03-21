package main

import (
	"fmt"
	"os"

	"server/internal/config"
	"server/internal/routes"
	"server/seed"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()
	config.ConnectDB()
	//TODO: Not ideal to run seeds on every startup, but this is a temporary measure until we have a better solution for managing test data.
	seed.Run(config.DB)

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
