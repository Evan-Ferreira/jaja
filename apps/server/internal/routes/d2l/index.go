package d2l

import (
	"github.com/gin-gonic/gin"
)

type authRequest struct {
	Cookies      map[string]any `json:"cookies"`
	LocalStorage map[string]any `json:"local_storage"`
}

func GetD2LRouter(router *gin.RouterGroup) *gin.RouterGroup {
	d2lGroup := router.Group("/d2l")
	{
		d2lGroup.POST("/auth", SaveCookiesAndLocalStorage)
	}
	return d2lGroup
}