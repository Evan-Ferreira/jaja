package d2lhandlers

import (
	"net/http"
	"strings"

	"server/internal/config"
	"server/internal/services"

	"github.com/gin-gonic/gin"
)

func GetWhoAmI(c *gin.Context) {

	// TODO: This is a temporary handler to test the D2L client. We will need to implement proper authentication and token management in the future
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	client := services.NewD2LClient(token, config.D2LBaseURL)

	user, err := client.GetWhoAmI()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
