package dev

import (
	"log"
	"net/http"
	"server/internal/services"

	"github.com/gin-gonic/gin"
)

func GetClaudeResponse(c *gin.Context) {
	claudeService, err := services.New()
	if err != nil {
		log.Printf("Failed to create Claude service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize AI service"})
		return
	}

	response, err := claudeService.Run(c.Request.Context(), services.ClaudeServiceConfig{
		Model:     "claude-sonnet-4-6",
		MaxTokens: 10000,
		Messages: []services.AnthropicMessage{
			{
				Role:    services.AnthropicRoleUser,
				Message: "What is the capital of France?",
			},
		},
	})
	if err != nil {
		log.Printf("Claude API request failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI request failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": response})
}
