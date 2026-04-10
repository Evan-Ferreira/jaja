package dev

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"server/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

type SaveAssignmentFilesRequest struct {
	AssignmentInstructionsRubric string `json:"assignment_instructions_rubric"`
}

// maxUploadSize is the maximum allowed file upload size (50 MB).
const maxUploadSize = 50 << 20

// allowedExtensions lists the file extensions accepted for assignment uploads.
var allowedExtensions = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".txt":  true,
	".rtf":  true,
	".odt":  true,
	".ppt":  true,
	".pptx": true,
	".xls":  true,
	".xlsx": true,
	".csv":  true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
}

// sanitizeFileName strips directory components and path traversal sequences
// so user-supplied names cannot escape the intended S3 prefix.
func sanitizeFileName(name string) string {
	// Take only the base filename, stripping any directory components.
	name = filepath.Base(name)

	// Reject path traversal attempts.
	if name == "." || name == ".." || name == "" {
		return ""
	}

	// Remove any remaining path separators (defense in depth).
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")

	return name
}

func SaveAssignmentFiles(c *gin.Context) {
	fileHeader, err := c.FormFile("assignment_instructions_rubric")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid file field"})
		return
	}

	// Enforce file size limit.
	if fileHeader.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file too large (max %d MB)", maxUploadSize>>20)})
		return
	}

	fileName := sanitizeFileName(c.PostForm("file_name"))
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing file name"})
		return
	}

	// Validate file extension.
	ext := strings.ToLower(filepath.Ext(fileName))
	if !allowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file type not allowed"})
		return
	}

	src, err := fileHeader.Open()

	if err != nil {
		log.Printf("Failed to open uploaded file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process uploaded file"})
		return
	}

	defer src.Close()

	_, err = config.S3BasicsBucket.S3Client.PutObject(c.Request.Context(), &s3.PutObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String(fileName),
		Body:   src,
	})

	if err != nil {
		log.Printf("Failed to upload file to S3: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("%s uploaded successfully", fileName), "file_name": fileName, "bucket_name": "test-bucket"})
}
