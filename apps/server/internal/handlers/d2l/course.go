package d2l

import (
	"log"
	"net/http"

	"server/internal/services"
	"server/seed"

	"github.com/gin-gonic/gin"
)

func GetCoursesAndAssignments(c *gin.Context) {
	// TODO: replace with real user ID from auth
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
