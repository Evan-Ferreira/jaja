package routes

import (
	d2lhandlers "server/internal/handlers/d2l"

	"github.com/gin-gonic/gin"
)

func RegisterD2LRoutes(rg *gin.RouterGroup) {
	d2l := rg.Group("/d2l")
	{
		d2l.POST("/auth", d2lhandlers.SaveAuth)
		d2l.GET("/users/whoami", d2lhandlers.GetWhoAmI)
		d2l.GET("/heartbeat", d2lhandlers.Heartbeat)
	}
}
