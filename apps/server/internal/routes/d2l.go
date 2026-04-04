package routes

import (
	"server/internal/handlers/d2l"

	"github.com/gin-gonic/gin"
)

func RegisterD2LRoutes(rg *gin.RouterGroup) {
	routes := rg.Group("/d2l")
	{
		routes.POST("/credentials", d2l.SaveCredentials)
		routes.GET("/courses", d2l.GetCoursesAndAssignments)
		routes.POST("/sync", d2l.SyncCoursesAndAssignments)
	}
}
