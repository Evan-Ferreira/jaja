package d2lhandlers

import (
	"net/http"

	"server/internal/services"
	"server/seed"

	"github.com/gin-gonic/gin"
)

func GetWhoAmI(c *gin.Context) {
	// TODO: replace TestUser with real authenticated user
	client, err := services.NewD2LClient(seed.TestUserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	user, err := client.GetWhoAmI()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
