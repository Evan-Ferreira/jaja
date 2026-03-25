package routes

import (
	"server/internal/handlers/dev"

	"github.com/gin-gonic/gin"
)

func RegisterDevRoutes(rg *gin.RouterGroup) {
	routes := rg.Group("/dev")
	{
		routes.POST("/assignment-files", dev.SaveAssignmentFiles)
		routes.POST("/complete-assignment", dev.CompleteAssignment)
	}
}