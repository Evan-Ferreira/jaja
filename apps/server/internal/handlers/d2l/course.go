package d2l

import (
	"log"
	"net/http"

	"server/internal/services"
	"server/seed"

	"github.com/gin-gonic/gin"
)

func GetCoursesAndAssignments(c *gin.Context) {
	client, err := services.NewD2LClient(seed.TestUserID)
	if err != nil {
		log.Printf("d2l: create client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create D2L client"})
		return
	}

	courses, err := client.LoadCoursesAndAssignments()
	if err != nil {
		log.Printf("d2l: load courses: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load courses and assignments"})
		return
	}

	c.JSON(http.StatusOK, courses)
}

func SyncCoursesAndAssignments(c *gin.Context) {
	client, err := services.NewD2LClient(seed.TestUserID)
	if err != nil {
		log.Printf("d2l: create client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create D2L client"})
		return
	}

	if err := client.SyncD2L(); err != nil {
		log.Printf("d2l: sync: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync courses and assignments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
