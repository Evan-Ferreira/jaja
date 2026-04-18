package dev

import (
	"net/http"
	"server/agent"
	agentRunner "server/agent/runner"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RunOrchestratedAgent(c *gin.Context) {
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

	response, err := agent.OrchestratedRunner.RunOrchestrated(c.Request.Context(), runInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": response})
}
