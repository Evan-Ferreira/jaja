package dev

import (
	"net/http"
	"server/agent"
	agentRunner "server/agent/runner"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RunAgentRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Prompt    string `json:"prompt" binding:"required"`
	UserID    string `json:"user_id"`
}

func RunAgent(c *gin.Context) {
	var req RunAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := uuid.Nil
	if req.UserID != "" {
		parsed, err := uuid.Parse(req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id: " + err.Error()})
			return
		}
		userID = parsed
	}

	runInput := agentRunner.RunInput{
		SessionID: req.SessionID,
		Prompt:    req.Prompt,
		UserID:    userID.String(),
	}

	response, err := agent.AgentRunner.Run(c.Request.Context(), runInput)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": response})
}
