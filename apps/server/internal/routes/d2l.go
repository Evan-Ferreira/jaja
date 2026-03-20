package routes

import (
	d2lhandlers "server/internal/handlers/d2l"

	"github.com/gin-gonic/gin"
)

func RegisterD2LRoutes(rg *gin.RouterGroup) {
	d2l := rg.Group("/d2l")
	{
		d2l.POST("/saveauth", d2lhandlers.SaveAuth)

	}
}
