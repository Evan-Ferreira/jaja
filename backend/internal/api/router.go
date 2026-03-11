package api

import (
	"backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

func NewRouter(d2lHandler *handlers.D2LHandler) *gin.Engine {
	router := gin.New()

	assignments := router.Group("/assignments")
	{
		assignments.GET("/", d2lHandler.Hello)
	}

	return router
}
