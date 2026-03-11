package d2l

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type D2LAuthRequest struct {
	Cookies      map[string]any `json:"cookies"`
	LocalStorage map[string]any `json:"local_storage"`
}

func SaveCookiesAndLocalStorage(c *gin.Context) {
	var req D2LAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("Cookies: ", req.Cookies)
	fmt.Println("Local Storage: ", req.LocalStorage)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cookies and local storage received",
	})
}