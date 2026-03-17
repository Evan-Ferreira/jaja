package d2lhandlers

import (
	"encoding/json"
	"log"
	"net/http"

	"server/internal/config"
	"server/internal/models"

	"github.com/gin-gonic/gin"
)

type authPayload struct {
	Cookies      map[string]string  `json:"cookies"`
	LocalStorage localStoragePayload `json:"local_storage"`
}

type localStoragePayload struct {
	FetchTokens    models.D2LFetchTokens `json:"D2L.Fetch.Tokens"`
	SessionExpired string      `json:"Session.Expired"`
	SessionLastAccessed string `json:"Session.LastAccessed"`
	SessionUserId  string      `json:"Session.UserId"`
	XsrfHitCodeSeed string    `json:"XSRF.HitCodeSeed"`
	XsrfToken      string      `json:"XSRF.Token"`
	PdfjsHistory   string      `json:"pdfjs.history"`
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

	// TODO: move away from hardcoded test user id
	cookieSession := models.D2LCookieSession{
		UserId:              models.TestUser.ID,
		Clck:                req.Cookies["_clck"],
		Clsk:                req.Cookies["_clsk"],
		D2LSameSiteCanaryA:  req.Cookies["d2lSameSiteCanaryA"],
		D2LSameSiteCanaryB:  req.Cookies["d2lSameSiteCanaryB"],
		D2LSecureSessionVal: req.Cookies["d2lSecureSessionVal"],
		D2LSessionVal:       req.Cookies["d2lSessionVal"],
	}

	fetchTokensJSON, err := json.Marshal(req.LocalStorage.FetchTokens)
	if err != nil {
		log.Printf("Failed to marshal fetch tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process fetch tokens"})
		return
	}

	// TODO: move away from hardcoded test user id
	localStorageSession := models.D2LLocalStorageSession{
		UserId:              models.TestUser.ID,
		D2LFetchTokens:      string(fetchTokensJSON),
		SessionExpired:      req.LocalStorage.SessionExpired,
		SessionLastAccessed: req.LocalStorage.SessionLastAccessed,
		SessionUserId:       req.LocalStorage.SessionUserId,
		XsrfHitCodeSeed:     req.LocalStorage.XsrfHitCodeSeed,
		XsrfToken:           req.LocalStorage.XsrfToken,
		PdfjsHistory:        req.LocalStorage.PdfjsHistory,
	}

	if result := config.DB.Create(&cookieSession); result.Error != nil {
		log.Printf("Failed to save cookie session to database: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cookie session to database"})
		return
	}

	if result := config.DB.Create(&localStorageSession); result.Error != nil {
		log.Printf("Failed to save local storage session to database: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save local storage session to database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
