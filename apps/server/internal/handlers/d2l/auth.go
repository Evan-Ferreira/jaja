package d2lhandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type authPayload struct {
	Cookies      map[string]string `json:"cookies"`
	LocalStorage localStoragePayload `json:"local_storage"`
}

type localStoragePayload struct {
	FetchTokens fetchTokens `json:"D2L.Fetch.Tokens"`
}

type fetchTokens struct {
	Wildcard tokenEntry `json:"*:*:*"`
}

type tokenEntry struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

func SaveAuth(c *gin.Context) {
	var req authPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := req.LocalStorage.FetchTokens.Wildcard.AccessToken
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing D2L bearer token"})
		return
	}

	// TODO: persist token to store layer
	c.JSON(http.StatusOK, gin.H{"success": true})
}
