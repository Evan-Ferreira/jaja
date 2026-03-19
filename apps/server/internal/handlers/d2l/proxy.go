package d2lhandlers

import (
	"net/http"

	"server/internal/services"
	"server/seed"

	"github.com/gin-gonic/gin"
)

func Proxy(c *gin.Context) {
	client, err := services.NewD2LClient(seed.TestUserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	client.Proxy(c)
}
