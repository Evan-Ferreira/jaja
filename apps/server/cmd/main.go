package main

import (
	"fmt"
	"os"

	"server/internal/api/routes/d2l"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	router := gin.Default()
	envErr := godotenv.Load()

	if envErr != nil {
		fmt.Println("Error loading .env file", envErr)
	}

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{os.Getenv("FRONTEND_URL")}
	router.Use(cors.New(config))

	api := router.Group("/api")
	{
		d2l.GetD2LRouter(api)
	}

	port := os.Getenv("PORT")
	fmt.Println("Starting server on http://localhost:" + port + "...")

	router.Run(":" + port)
}
