package routes

import (
	"server/internal/handlers/dev"

	"github.com/gin-gonic/gin"
)

func RegisterDevRoutes(rg *gin.RouterGroup) {
	routes := rg.Group("/dev")
	{
		routes.POST("/assignment-files", dev.SaveAssignmentFiles)
		routes.POST("/claude", dev.GetClaudeResponse)
		// routes.POST("/complete-assignment", dev.CompleteAssignment)
		routes.POST("/update-content", dev.UpdateContent)
		routes.POST("/presigned-url", dev.GeneratePresignedURL)
		routes.POST("/run-agent", dev.RunAgent)
		routes.POST("/run-orchestrated-agent", dev.RunOrchestratedAgent)
		routes.POST("/run-claude", dev.RunClaude)
	}
}
