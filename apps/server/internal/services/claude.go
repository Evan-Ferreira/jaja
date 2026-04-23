package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)
type ClaudeService struct {
	Client *anthropic.Client
}

type AnthropicMessageRole string

const (
	AnthropicRoleUser      AnthropicMessageRole = "user"
	AnthropicRoleAssistant AnthropicMessageRole = "assistant"
)

type AnthropicMessage struct {
	Role          AnthropicMessageRole
	Message       string
	ContentBlocks []anthropic.BetaContentBlockParamUnion
}

type DocumentType string

const (
	DocumentTypePDF  DocumentType = "application/pdf"
	DocumentTypePNG  DocumentType = "image/png"
	DocumentTypeJPEG DocumentType = "image/jpeg"
	DocumentTypeWebP DocumentType = "image/webp"
	DocumentTypeGIF  DocumentType = "image/gif"
)

type PresignedDocument struct {
	URL  string
	Type DocumentType
}


type ClaudeServiceConfig struct {
	Model     anthropic.Model
	MaxTokens int64
	Messages  []AnthropicMessage
	Documents []PresignedDocument
	Tools     *[]anthropic.BetaToolUnionParam
	Betas     *[]anthropic.AnthropicBeta
	Skills    *[]anthropic.BetaSkillParams
}


func newClient() (*anthropic.Client, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")

	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey), 
	)

	return &client, nil
}



func New() (*ClaudeService, error){
	client, err := newClient()

	if err != nil {
		return nil, err
	}

	return &ClaudeService{
		client,
	}, nil
}

func (claudeService *ClaudeService) Run(ctx context.Context, config ClaudeServiceConfig) (*anthropic.BetaMessage, error) {
	messages, err := parseMessagesParam(config.Messages)
	if err != nil {
		return nil, err
	}

	params := anthropic.BetaMessageNewParams{
		Model:     config.Model,
		MaxTokens: config.MaxTokens,
		Messages:  messages,
	}

	if config.Skills != nil {
		params.Container = anthropic.BetaMessageNewParamsContainerUnion{
			OfContainers: &anthropic.BetaContainerParams{
				Skills: *config.Skills,
			},
		}
		// in order to use skills, we must add the following betas
		params.Betas = append(params.Betas,
			anthropic.AnthropicBetaSkills2025_10_02,
			anthropic.AnthropicBeta("code-execution-2025-08-25"),
			anthropic.AnthropicBetaFilesAPI2025_04_14,
		)
		// in order to use the skills, we must add code execution as tool
		params.Tools = append(params.Tools, anthropic.BetaToolUnionParam{
			OfCodeExecutionTool20250825: &anthropic.BetaCodeExecutionTool20250825Param{},
		})
	}

	if config.Tools != nil {
		params.Tools = append(params.Tools, *config.Tools...)
	}

	if config.Documents != nil {
		docBlocks, err := buildDocumentBlocks(config.Documents)
		if err != nil {
			return nil, fmt.Errorf("error building document blocks: %w", err)
		}
		params.Messages = append(params.Messages, anthropic.BetaMessageParam{
			Role:    anthropic.BetaMessageParamRoleUser,
			Content: docBlocks,
		})
	}
	if config.Betas != nil {
		params.Betas = *config.Betas
	}

	response, err := claudeService.Client.Beta.Messages.New(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("claude response: %w", err)
	}

	return response, nil
}

func parseMessagesParam(messages []AnthropicMessage) ([]anthropic.BetaMessageParam, error) {
	res := make([]anthropic.BetaMessageParam, 0, len(messages))

	for _, m := range messages {
		var role anthropic.BetaMessageParamRole
		switch m.Role {
		case AnthropicRoleUser:
			role = anthropic.BetaMessageParamRoleUser
		case AnthropicRoleAssistant:
			role = anthropic.BetaMessageParamRoleAssistant
		default:
			return nil, fmt.Errorf("unknown message role: %q", m.Role)
		}

		content := m.ContentBlocks
		if m.Message != "" {
			content = append(content, anthropic.NewBetaTextBlock(m.Message))
		}

		res = append(res, anthropic.BetaMessageParam{Role: role, Content: content})
	}

	return res, nil
}

type DownloadedFile struct {
	Filename string
	MimeType string
	Data     []byte
}

func (cs *ClaudeService) GetFilesFromResponse(ctx context.Context, response *anthropic.BetaMessage) ([]DownloadedFile, error) {
	fileIds := ExtractFileIDs(response)
	files, err := cs.DownloadFiles(ctx, fileIds)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func ExtractFileIDs(response *anthropic.BetaMessage) []string {
	var fileIDs []string
	for _, item := range response.Content {
		switch variant := item.AsAny().(type) {
		case anthropic.BetaBashCodeExecutionToolResultBlock:
			for _, file := range variant.Content.Content {
				if file.FileID != "" {
					fileIDs = append(fileIDs, file.FileID)
				}
			}
		case anthropic.BetaContainerUploadBlock:
			if variant.FileID != "" {
				fileIDs = append(fileIDs, variant.FileID)
			}
		}
	}
	return fileIDs
}

func (cs *ClaudeService) DownloadFiles(ctx context.Context, fileIDs []string) ([]DownloadedFile, error) {
	filesBeta := []anthropic.AnthropicBeta{anthropic.AnthropicBetaFilesAPI2025_04_14}
	var files []DownloadedFile

	for _, fid := range fileIDs {
		metadata, err := cs.Client.Beta.Files.GetMetadata(ctx, fid, anthropic.BetaFileGetMetadataParams{
			Betas: filesBeta,
		})

		if err != nil {
			return nil, fmt.Errorf("get metadata for file %s: %w", fid, err)
		}

		resp, err := cs.Client.Beta.Files.Download(ctx, fid, anthropic.BetaFileDownloadParams{
			Betas: filesBeta,
		})
		if err != nil {
			return nil, fmt.Errorf("download file %s: %w", fid, err)
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read file %s body: %w", fid, err)
		}

		files = append(files, DownloadedFile{
			Filename: metadata.Filename,
			MimeType: metadata.MimeType,
			Data:     data,
		})
	}
	return files, nil
}

func InferDocumentType(rawURL string) DocumentType {
	path := rawURL
	if i := strings.Index(rawURL, "?"); i >= 0 {
		path = rawURL[:i]
	}
	path = strings.ToLower(path)
	switch {
		case strings.HasSuffix(path, ".png"):
			return DocumentTypePNG
		case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
			return DocumentTypeJPEG
		case strings.HasSuffix(path, ".webp"):
			return DocumentTypeWebP
		case strings.HasSuffix(path, ".gif"):
			return DocumentTypeGIF
		default:
			return DocumentTypePDF
	}
}
// converts a list of presigned URLs to the appropriate
// claude content blocks (BetaDocumentBlock for PDFs, BetaImageBlock for images).
func buildDocumentBlocks(docs []PresignedDocument) ([]anthropic.BetaContentBlockParamUnion, error) {
	blocks := make([]anthropic.BetaContentBlockParamUnion, 0, len(docs))
	for _, doc := range docs {
		switch doc.Type {
		case DocumentTypePDF:
			blocks = append(blocks, anthropic.NewBetaDocumentBlock(anthropic.BetaURLPDFSourceParam{URL: doc.URL}))
		case DocumentTypePNG, DocumentTypeJPEG, DocumentTypeWebP, DocumentTypeGIF:
			blocks = append(blocks, anthropic.NewBetaImageBlock(anthropic.BetaURLImageSourceParam{URL: doc.URL}))
		default:
			return nil, fmt.Errorf("unsupported document type: %q", doc.Type)
		}
	}
	return blocks, nil
}
