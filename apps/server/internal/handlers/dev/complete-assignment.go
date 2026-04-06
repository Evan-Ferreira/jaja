package dev

type completeAssignmentReq struct {
	AssignmentName string `json:"assignment_name"`
	AssignmentFile string `json:"assignment_file"`
	Prompt string `json:"prompt"`
}

// func CompleteAssignment(c *gin.Context){
// 	var req completeAssignmentReq

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	anthropicClient, err := agent.New()

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	presignedURL, err := config.S3BasicsBucket.GeneratePresignedUrl(c.Request.Context(), "test-bucket", req.AssignmentFile, 0)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	response, err := anthropicClient.Run(c.Request.Context(), "claude-sonnet-4-6", req.Prompt, &presignedURL)
	
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"response": response})
// }