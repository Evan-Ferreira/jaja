package dev

import (
	"net/http"
	"server/internal/config"

	"github.com/gin-gonic/gin"
)

type GeneratePresignedURLRequest struct {
	BucketName    string `json:"bucket_name" binding:"required"`
	FileKey       string `json:"file_key" binding:"required"`
	ExpireSeconds int    `json:"expire_seconds"`
}

func GeneratePresignedURL(c *gin.Context) {
	var req GeneratePresignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, err := config.S3BasicsBucket.GeneratePresignedUrl(c.Request.Context(), req.BucketName, req.FileKey, req.ExpireSeconds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}
