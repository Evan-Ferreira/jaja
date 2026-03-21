package d2l

import (
	"log"
	"net/http"

	"server/internal/database"
	"server/internal/models"

	"github.com/gin-gonic/gin"
)

type D2LAuthRequest struct {
	Cookies      models.D2LCookieSession `json:"cookies"`
	LocalStorage models.D2LLocalStorageSession `json:"local_storage"`
}

func SaveCookiesAndLocalStorage(c *gin.Context) {
	var req D2LAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing D2L cookies or local storage field(s)"})
		return
	}

	// TODO: move away from hardcoded test user id
	userCookieSession := models.D2LCookieSession{
		UserId: models.TestUser.ID,
		Clck: req.Cookies.Clck,
		Clsk: req.Cookies.Clsk,
		D2LSameSiteCanaryA: req.Cookies.D2LSameSiteCanaryA,
		D2LSameSiteCanaryB: req.Cookies.D2LSameSiteCanaryB,
		D2LSecureSessionVal: req.Cookies.D2LSecureSessionVal,
		D2LSessionVal: req.Cookies.D2LSessionVal,
	}

	// TODO: move away from hardcoded test user id
	userLocalStorageSession := models.D2LLocalStorageSession{
		UserId: models.TestUser.ID,
		D2LFetchTokens: req.LocalStorage.D2LFetchTokens,
		SessionExpired: req.LocalStorage.SessionExpired,
		SessionLastAccessed: req.LocalStorage.SessionLastAccessed,
		SessionUserId: req.LocalStorage.SessionUserId,
		XsrfHitCodeSeed: req.LocalStorage.XsrfHitCodeSeed,
		XsrfToken: req.LocalStorage.XsrfToken,
		PdfjsHistory: req.LocalStorage.PdfjsHistory,
	}

	result := database.DBClient.Create(&userCookieSession)
	if result.Error != nil {
		log.Printf("Failed save cookie session to database: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed save cookie session to database"})
		return
	}

	result = database.DBClient.Create(&userLocalStorageSession)
	if result.Error != nil {
		log.Printf("Failed save local storage session to database: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed save local storage session to database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cookies and local storage received",
	})
}