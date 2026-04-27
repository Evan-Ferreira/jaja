package routes

import (
	"server/internal/mockapi/handlers/anthropic"

	"github.com/gin-gonic/gin"
)

func RegisterAnthropicRoutes(rg *gin.RouterGroup) {
	routes := rg.Group("/v1")
	{
		routes.POST("/messages", anthropic.HandleMessages)
		routes.GET("/files/:id", anthropic.HandleFileMetadata)
		routes.GET("/files/:id/content", anthropic.HandleFileContent)
	}
}
