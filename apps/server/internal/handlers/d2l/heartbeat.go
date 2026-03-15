package d2lhandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Heartbeat(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"status": "Hi, I'm alive 22222!"})
}
