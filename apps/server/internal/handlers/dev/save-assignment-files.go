package dev

import (
	"fmt"
	"net/http"

	"server/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

type SaveAssignmentFilesRequest struct {
	AssignmentInstructionsRubric string `json:"assignment_instructions_rubric"`
}

func SaveAssignmentFiles(c *gin.Context) {
	fileHeader, err := c.FormFile("assignment_instructions_rubric")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fileName := c.PostForm("file_name")

	src, err := fileHeader.Open()

	_, err = config.S3Client.PutObject(c.Request.Context(), &s3.PutObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String(fileName),
		Body:   src,
	})

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("%s uploaded successfully", fileName), "file_name": fileName, "bucket_name": "test-bucket"})
}