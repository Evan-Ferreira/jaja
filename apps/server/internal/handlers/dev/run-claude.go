package dev

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"server/internal/services"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/gin-gonic/gin"
)

type RunClaudeRequest struct {
	Prompt      string   `json:"prompt" binding:"required"`
	DocumentURL string   `json:"document_url"`
	Model       string   `json:"model"`
	MaxTokens   int64    `json:"max_tokens"`
}

func RunClaude(c *gin.Context) {
	var req RunClaudeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Model == "" {
		req.Model = "claude-sonnet-4-6"
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 20000
	}

	claudeService, err := services.New()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	config := services.ClaudeServiceConfig{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		Messages: []services.AnthropicMessage{
			{
				Role:    services.AnthropicRoleUser,
				Message: req.Prompt,
			},
		},
		Skills: &[]anthropic.BetaSkillParams{
			{
				SkillID: "docx",
				Type:    anthropic.BetaSkillParamsTypeAnthropic,
				Version: anthropic.String("latest"),
			},
		},
	}

	if req.DocumentURL != "" {
		config.Documents = []services.PresignedDocument{
			{
				URL:  req.DocumentURL,
				Type: inferDocumentType(req.DocumentURL),
			},
		}
	}

	response, err := claudeService.Run(c.Request.Context(), config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fileIDs := services.ExtractFileIDs(response)
	var downloadedFiles []gin.H
	if len(fileIDs) > 0 {
		files, err := claudeService.DownloadFiles(c.Request.Context(), fileIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		outDir := "tmp/claude-files"
		if err := os.MkdirAll(outDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("create output dir: %s", err)})
			return
		}

		for _, f := range files {
			outPath := filepath.Join(outDir, f.Filename)
			if err := os.WriteFile(outPath, f.Data, 0644); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("write file %s: %s", f.Filename, err)})
				return
			}
			fmt.Printf("[RunClaude] Saved %s (%d bytes) to %s\n", f.Filename, len(f.Data), outPath)

			downloadedFiles = append(downloadedFiles, gin.H{
				"filename":  f.Filename,
				"mime_type": f.MimeType,
				"size":      len(f.Data),
				"path":      outPath,
				"data_b64":  base64.StdEncoding.EncodeToString(f.Data),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"response": response, "files": downloadedFiles})
}

func inferDocumentType(rawURL string) services.DocumentType {
	path, _, _ := strings.Cut(rawURL, "?")
	path = strings.ToLower(path)
	switch {
	case strings.HasSuffix(path, ".png"):
		return services.DocumentTypePNG
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		return services.DocumentTypeJPEG
	case strings.HasSuffix(path, ".webp"):
		return services.DocumentTypeWebP
	case strings.HasSuffix(path, ".gif"):
		return services.DocumentTypeGIF
	default:
		return services.DocumentTypePDF
	}
}
