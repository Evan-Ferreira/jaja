package main

import (
	"os"

	"server/internal/mockapi/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api := router.Group("/")
	{
		routes.RegisterAnthropicRoutes(api)
	}

	router.Run(":" + os.Getenv("MOCKAPI_PORT"))
}
