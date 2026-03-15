package d2l

import (
	"log"
	"net/http"

	"server/internal/models"

	"server/internal"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type D2LAuthRequest struct {
	Cookies      models.D2LCookies `json:"cookies"`
	LocalStorage models.D2LLocalStorage `json:"local_storage"`
}

func SaveCookiesAndLocalStorage(c *gin.Context) {
	var req D2LAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := internal.DB.FirstOrCreate(&models.User{
		ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
	})

	cookies := models.D2LCookies{
		UserId: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Clck: req.Cookies.Clck,
		Clsk: req.Cookies.Clsk,
		D2LSameSiteCanaryA: req.Cookies.D2LSameSiteCanaryA,
		D2LSameSiteCanaryB: req.Cookies.D2LSameSiteCanaryB,
		D2LSecureSessionVal: req.Cookies.D2LSecureSessionVal,
		D2LSessionVal: req.Cookies.D2LSessionVal,
	}

	localStorage := models.D2LLocalStorage{
		UserId: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		D2LFetchTokens: req.LocalStorage.D2LFetchTokens,
		SessionExpired: req.LocalStorage.SessionExpired,
		SessionLastAccessed: req.LocalStorage.SessionLastAccessed,
		SessionUserId: req.LocalStorage.SessionUserId,
		XsrfHitCodeSeed: req.LocalStorage.XsrfHitCodeSeed,
		XsrfToken: req.LocalStorage.XsrfToken,
		PdfjsHistory: req.LocalStorage.PdfjsHistory,
	}

	result = internal.DB.Create(&cookies)
	if result.Error != nil {
		log.Println("Failed to create cookies: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	result = internal.DB.Create(&localStorage)
	if result.Error != nil {
		log.Println("Failed to create local storage: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cookies and local storage received",
	})
}