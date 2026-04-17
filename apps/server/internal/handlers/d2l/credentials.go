package d2l

import (
	"log"
	"net/http"

	"server/internal/database"
	"server/internal/models"
	"server/seed"

	"github.com/gin-gonic/gin"
)

type authPayload struct {
	Cookies      map[string]string   `json:"cookies"`
	LocalStorage localStoragePayload `json:"local_storage"`
}

type fetchTokensPayload struct {
	Wildcard struct {
		AccessToken string `json:"access_token"`
		ExpiresAt   int64  `json:"expires_at"`
	} `json:"*:*:*"`
}

type localStoragePayload struct {
	FetchTokens         fetchTokensPayload `json:"D2L.Fetch.Tokens"`
	SessionExpired      string             `json:"Session.Expired"`
	SessionLastAccessed string             `json:"Session.LastAccessed"`
	SessionUserId       string             `json:"Session.UserId"`
	XsrfHitCodeSeed     string             `json:"XSRF.HitCodeSeed"`
	XsrfToken           string             `json:"XSRF.Token"`
	PdfjsHistory        string             `json:"pdfjs.history"`
}


func SaveCredentials(c *gin.Context) {
	var req authPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Failed to bind credentials JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	token := req.LocalStorage.FetchTokens.Wildcard.AccessToken
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing D2L bearer token"})
		return
	}

	// TODO: move away from hardcoded test user id
	cookieSession := models.D2LCookieSession{
		UserId:              seed.TestUserID,
		Clck:                req.Cookies["_clck"],
		Clsk:                req.Cookies["_clsk"],
		D2LSameSiteCanaryA:  req.Cookies["d2lSameSiteCanaryA"],
		D2LSameSiteCanaryB:  req.Cookies["d2lSameSiteCanaryB"],
		D2LSecureSessionVal: req.Cookies["d2lSecureSessionVal"],
		D2LSessionVal:       req.Cookies["d2lSessionVal"],
	}

	// TODO: move away from hardcoded test user id
	localStorageSession := models.D2LLocalStorageSession{
		UserId:              seed.TestUserID,
		FetchAccessToken:    req.LocalStorage.FetchTokens.Wildcard.AccessToken,
		FetchExpiresAt:      req.LocalStorage.FetchTokens.Wildcard.ExpiresAt,
		SessionExpired:      req.LocalStorage.SessionExpired,
		SessionLastAccessed: req.LocalStorage.SessionLastAccessed,
		SessionUserId:       req.LocalStorage.SessionUserId,
		XsrfHitCodeSeed:     req.LocalStorage.XsrfHitCodeSeed,
		XsrfToken:           req.LocalStorage.XsrfToken,
		PdfjsHistory:        req.LocalStorage.PdfjsHistory,
	}

	if result := database.DBClient.Create(&cookieSession); result.Error != nil {
		log.Printf("Failed to save cookie session to database: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cookie session to database"})
		return
	}

	if result := database.DBClient.Create(&localStorageSession); result.Error != nil {
		log.Printf("Failed to save local storage session to database: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save local storage session to database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
