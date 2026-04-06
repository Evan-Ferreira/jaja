package dev

import (
	"context"
	"log"
	"net/http"

	"server/internal/services"
	"server/seed"

	"github.com/gin-gonic/gin"
)

// testOrgUnitID is the D2L org unit ID used for manual content sync testing.
const testOrgUnitID = 997744

func UpdateContent(c *gin.Context) {
	client, err := services.NewD2LClient(seed.TestUserID)
	if err != nil {
		log.Printf("dev: update content: create client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create D2L client"})
		return
	}

	client.UpdateContent(context.Background(), testOrgUnitID)

	c.JSON(http.StatusOK, gin.H{"success": true, "org_unit_id": testOrgUnitID})
}
