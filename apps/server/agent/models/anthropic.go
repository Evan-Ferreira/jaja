// Package models implements the ADK [model.LLM] interface for Anthropic Claude
// models. It mirrors google/adk-java's com.google.adk.models.Claude, translating
// ADK LLMRequests into Anthropic MessageNewParams and back.
//
// Streaming and live connections are not currently supported.
package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/joho/godotenv"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

type AnthropicModel struct {
	Client    *anthropic.Client
	ModelName anthropic.Model
	MaxTokens int64
}

const DefaultMaxTokens int64 = 8192

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

type AnthropicServiceConfig struct {
	Model     anthropic.Model
	MaxTokens int64
	Messages  []AnthropicMessage
	Documents []PresignedDocument
	Tools     *[]anthropic.BetaToolUnionParam
	Betas     *[]anthropic.AnthropicBeta
	Skills    *[]anthropic.BetaSkillParams
}

func New(modelName anthropic.Model) (*AnthropicModel, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}

	baseURL := os.Getenv("ANTHROPIC_API_URL")
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	
	client := anthropic.NewClient(opts...)
	return &AnthropicModel{
		Client:    &client,
		ModelName: modelName,
		MaxTokens: DefaultMaxTokens,
	}, nil
}

func (anthropicModel *AnthropicModel) Run(ctx context.Context, config AnthropicServiceConfig) (*anthropic.BetaMessage, error) {
	messages, err := anthropicModel.parseMessagesParam(config.Messages)
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

	response, err := anthropicModel.Client.Beta.Messages.New(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("claude response: %w", err)
	}

	return response, nil
}

func (anthropicModel *AnthropicModel) parseMessagesParam(messages []AnthropicMessage) ([]anthropic.BetaMessageParam, error) {
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

func (anthropicModel *AnthropicModel) GetFilesFromResponse(ctx context.Context, response *anthropic.BetaMessage) ([]DownloadedFile, error) {
	fileIds := ExtractFileIDs(response)
	files, err := anthropicModel.DownloadFiles(ctx, fileIds)
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

func (anthropicModel *AnthropicModel) DownloadFiles(ctx context.Context, fileIDs []string) ([]DownloadedFile, error) {
	filesBeta := []anthropic.AnthropicBeta{anthropic.AnthropicBetaFilesAPI2025_04_14}
	var files []DownloadedFile

	for _, fid := range fileIDs {
		metadata, err := anthropicModel.Client.Beta.Files.GetMetadata(ctx, fid, anthropic.BetaFileGetMetadataParams{
			Betas: filesBeta,
		})

		if err != nil {
			return nil, fmt.Errorf("get metadata for file %s: %w", fid, err)
		}

		resp, err := anthropicModel.Client.Beta.Files.Download(ctx, fid, anthropic.BetaFileDownloadParams{
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

// generateStream streams responses from the Anthropic API, yielding text deltas
// as partial responses and a final complete response with TurnComplete set to true.
func (anthropicModel *AnthropicModel) generateStream(ctx context.Context, req *model.LLMRequest) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		params, err := anthropicModel.buildParams(req)
		if err != nil {
			yield(nil, fmt.Errorf("failed to build Anthropic params: %w", err))
			return
		}

		stream := anthropicModel.Client.Messages.NewStreaming(ctx, params)
		accumulated := anthropic.Message{}

		for stream.Next() {
			event := stream.Current()

			if err := accumulated.Accumulate(event); err != nil {
				yield(nil, fmt.Errorf("failed to accumulate stream event: %w", err))
				return
			}

			// Yield text deltas as partial responses.
			if delta, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
				if text, ok := delta.Delta.AsAny().(anthropic.TextDelta); ok {
					if !yield(&model.LLMResponse{
						Content: &genai.Content{
							Role:  "model",
							Parts: []*genai.Part{{Text: text.Text}},
						},
						TurnComplete: false,
					}, nil) {
						return
					}
				}
			}
		}

		if err := stream.Err(); err != nil {
			yield(nil, fmt.Errorf("stream error: %w", err))
			return
		}

		// Yield the final complete response.
		final := convertResponse(&accumulated)
		final.TurnComplete = true
		yield(final, nil)
	}
}

func (anthropicModel *AnthropicModel) Name() string { return anthropicModel.ModelName }

// GenerateContent implements model.LLM. The `stream` argument is ignored —
func (anthropicModel *AnthropicModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	if stream {
		return anthropicModel.generateStream(ctx, req)
	}
	return func(yield func(*model.LLMResponse, error) bool) {
		params, err := anthropicModel.buildParams(req)
		if err != nil {
			yield(nil, fmt.Errorf("build anthropic params: %w", err))
			return
		}

		msg, err := anthropicModel.Client.Messages.New(ctx, params)
		if err != nil {
			yield(nil, fmt.Errorf("anthropic messages.new: %w", err))
			return
		}
		yield(convertResponse(msg), nil)
	}
}

func (anthropicModel *AnthropicModel) buildParams(req *model.LLMRequest) (anthropic.MessageNewParams, error) {
	modelName := anthropicModel.ModelName
	if req.Model != "" {
		modelName = req.Model
	}

	params := anthropic.MessageNewParams{
		Model:     modelName,
		MaxTokens: anthropicModel.MaxTokens,
	}

	msgs, err := convertMessages(req.Contents)
	if err != nil {
		return params, err
	}
	params.Messages = msgs

	if req.Config == nil {
		return params, nil
	}

	if sys := extractSystemText(req.Config.SystemInstruction); sys != "" {
		params.System = []anthropic.TextBlockParam{{Text: sys}}
	}

	if req.Config.MaxOutputTokens > 0 {
		params.MaxTokens = int64(req.Config.MaxOutputTokens)
	}
	if req.Config.Temperature != nil {
		params.Temperature = param.NewOpt(float64(*req.Config.Temperature))
	}
	if req.Config.TopP != nil {
		params.TopP = param.NewOpt(float64(*req.Config.TopP))
	}
	if req.Config.TopK != nil {
		params.TopK = param.NewOpt(int64(*req.Config.TopK))
	}
	if len(req.Config.StopSequences) > 0 {
		params.StopSequences = req.Config.StopSequences
	}

	if tools := convertTools(req.Config.Tools); len(tools) > 0 {
		params.Tools = tools
		params.ToolChoice = anthropic.ToolChoiceUnionParam{
			OfAuto: &anthropic.ToolChoiceAutoParam{
				DisableParallelToolUse: param.NewOpt(true),
			},
		}
	}

	return params, nil
}

// extractSystemText joins the text parts of a system instruction with newlines,
// mirroring Java's `parts.filter(text).collect(joining("\n"))`.
func extractSystemText(content *genai.Content) string {
	if content == nil {
		return ""
	}
	var parts []string
	for _, p := range content.Parts {
		if p == nil || p.Text == "" {
			continue
		}
		parts = append(parts, p.Text)
	}
	return strings.Join(parts, "\n")
}

func convertMessages(contents []*genai.Content) ([]anthropic.MessageParam, error) {
	msgs := make([]anthropic.MessageParam, 0, len(contents))
	for _, c := range contents {
		if c == nil {
			continue
		}
		blocks, err := convertParts(c.Parts)
		if err != nil {
			return nil, fmt.Errorf("convert parts for role %q: %w", c.Role, err)
		}
		if len(blocks) == 0 {
			continue
		}
		msgs = append(msgs, anthropic.MessageParam{
			Role:    convertRole(c.Role),
			Content: blocks,
		})
	}
	return msgs, nil
}

func convertRole(role string) anthropic.MessageParamRole {
	if role == "model" || role == "assistant" {
		return anthropic.MessageParamRoleAssistant
	}
	return anthropic.MessageParamRoleUser
}

func convertParts(parts []*genai.Part) ([]anthropic.ContentBlockParamUnion, error) {
	var blocks []anthropic.ContentBlockParamUnion
	for _, p := range parts {
		if p == nil {
			continue
		}
		switch {
		case p.Text != "":
			blocks = append(blocks, anthropic.NewTextBlock(p.Text))

		case p.FunctionCall != nil:
			fc := p.FunctionCall
			args := fc.Args
			if args == nil {
				args = map[string]any{}
			}
			blocks = append(blocks, anthropic.NewToolUseBlock(fc.ID, args, fc.Name))

		case p.FunctionResponse != nil:
			fr := p.FunctionResponse
			content, err := functionResponseContent(fr.Response)
			if err != nil {
				return nil, fmt.Errorf("marshal function response %q: %w", fr.Name, err)
			}
			blocks = append(blocks, anthropic.NewToolResultBlock(fr.ID, content, false))

		case p.InlineData != nil:
			blob := p.InlineData
			blocks = append(blocks, anthropic.NewImageBlockBase64(blob.MIMEType, string(blob.Data)))

		case p.FileData != nil:
			fd := p.FileData
			switch {
			case strings.HasPrefix(fd.MIMEType, "image/"):
				blocks = append(blocks, anthropic.NewImageBlock(anthropic.URLImageSourceParam{URL: fd.FileURI}))
			default:
				blocks = append(blocks, anthropic.NewDocumentBlock(anthropic.URLPDFSourceParam{URL: fd.FileURI}))
			}
		}
	}
	return blocks, nil
}

// functionResponseContent mirrors the Java Claude.java behavior: prefer a
// conventional "result" key, then "output", otherwise serialize the whole map.
func functionResponseContent(resp map[string]any) (string, error) {
	if len(resp) == 0 {
		return "", nil
	}
	for _, key := range []string{"result", "output"} {
		if v, ok := resp[key]; ok {
			return stringify(v), nil
		}
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func stringify(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// convertTools translates genai tool declarations into Anthropic ToolUnionParam
// values. Only function declarations are forwarded — other genai tool kinds
// (GoogleSearch, Retrieval, etc.) have no Claude equivalent and are ignored.
func convertTools(tools []*genai.Tool) []anthropic.ToolUnionParam {
	var out []anthropic.ToolUnionParam
	for _, t := range tools {
		if t == nil {
			continue
		}
		for _, fd := range t.FunctionDeclarations {
			if fd == nil || fd.Name == "" {
				continue
			}
			tp := anthropic.ToolParam{
				Name:        fd.Name,
				InputSchema: buildInputSchemaForDecl(fd.Parameters, fd.ParametersJsonSchema),
			}
			if fd.Description != "" {
				tp.Description = param.NewOpt(fd.Description)
			}
			out = append(out, anthropic.ToolUnionParam{OfTool: &tp})
		}
	}
	return out
}

// buildInputSchemaForDecl prefers the structured genai.Schema; falls back to the
// raw ParametersJsonSchema that functiontool.New() populates instead.
func buildInputSchemaForDecl(s *genai.Schema, raw any) anthropic.ToolInputSchemaParam {
	if s != nil {
		return buildInputSchema(s)
	}
	if raw == nil {
		return anthropic.ToolInputSchemaParam{Properties: map[string]any{}}
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return anthropic.ToolInputSchemaParam{Properties: map[string]any{}}
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return anthropic.ToolInputSchemaParam{Properties: map[string]any{}}
	}
	schema := anthropic.ToolInputSchemaParam{Properties: map[string]any{}}
	if props, ok := m["properties"]; ok {
		schema.Properties = props
	}
	if req, ok := m["required"].([]any); ok {
		required := make([]string, 0, len(req))
		for _, r := range req {
			if str, ok := r.(string); ok {
				required = append(required, str)
			}
		}
		schema.Required = required
	}
	return schema
}

func buildInputSchema(s *genai.Schema) anthropic.ToolInputSchemaParam {
	schema := anthropic.ToolInputSchemaParam{Properties: map[string]any{}}
	if s == nil {
		return schema
	}
	if len(s.Properties) > 0 {
		props := make(map[string]any, len(s.Properties))
		for k, v := range s.Properties {
			props[k] = schemaToMap(v)
		}
		schema.Properties = props
	}
	if len(s.Required) > 0 {
		schema.Required = s.Required
	}
	return schema
}

// schemaToMap converts a genai.Schema into a JSON Schema-shaped map, normalizing
// uppercase OpenAPI-style types ("STRING") into the lowercase variants Anthropic
// expects ("string"). Mirrors Java Claude.updateTypeString, extended to walk the
// full schema tree.
func schemaToMap(s *genai.Schema) map[string]any {
	if s == nil {
		return nil
	}
	out := map[string]any{}
	if s.Type != "" {
		out["type"] = strings.ToLower(string(s.Type))
	}
	if s.Description != "" {
		out["description"] = s.Description
	}
	if len(s.Enum) > 0 {
		out["enum"] = s.Enum
	}
	if s.Format != "" {
		out["format"] = s.Format
	}
	if s.Default != nil {
		out["default"] = s.Default
	}
	if s.Nullable != nil {
		out["nullable"] = *s.Nullable
	}
	if s.Pattern != "" {
		out["pattern"] = s.Pattern
	}
	if s.Items != nil {
		out["items"] = schemaToMap(s.Items)
	}
	if len(s.Properties) > 0 {
		props := make(map[string]any, len(s.Properties))
		for k, v := range s.Properties {
			props[k] = schemaToMap(v)
		}
		out["properties"] = props
	}
	if len(s.Required) > 0 {
		out["required"] = s.Required
	}
	if len(s.AnyOf) > 0 {
		anyOf := make([]any, 0, len(s.AnyOf))
		for _, sub := range s.AnyOf {
			anyOf = append(anyOf, schemaToMap(sub))
		}
		out["anyOf"] = anyOf
	}
	return out
}

func convertResponse(msg *anthropic.Message) *model.LLMResponse {
	return &model.LLMResponse{
		Content: &genai.Content{
			Role:  "model",
			Parts: convertResponseBlocks(msg.Content),
		},
		FinishReason: convertStopReason(msg.StopReason),
		TurnComplete: true,
		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
			PromptTokenCount:     int32(msg.Usage.InputTokens),
			CandidatesTokenCount: int32(msg.Usage.OutputTokens),
			TotalTokenCount:      int32(msg.Usage.InputTokens + msg.Usage.OutputTokens),
		},
	}
}

func convertResponseBlocks(blocks []anthropic.ContentBlockUnion) []*genai.Part {
	var parts []*genai.Part
	for _, b := range blocks {
		switch b.Type {
		case "text":
			parts = append(parts, &genai.Part{Text: b.Text})
		case "tool_use":
			args := map[string]any{}
			if len(b.Input) > 0 {
				_ = json.Unmarshal(b.Input, &args)
			}
			parts = append(parts, &genai.Part{
				FunctionCall: &genai.FunctionCall{
					ID:   b.ID,
					Name: b.Name,
					Args: args,
				},
			})
		}
	}
	return parts
}

func convertStopReason(reason anthropic.StopReason) genai.FinishReason {
	switch reason {
	case anthropic.StopReasonEndTurn,
		anthropic.StopReasonStopSequence,
		anthropic.StopReasonToolUse,
		anthropic.StopReasonPauseTurn:
		return genai.FinishReasonStop
	case anthropic.StopReasonMaxTokens:
		return genai.FinishReasonMaxTokens
	case anthropic.StopReasonRefusal:
		return genai.FinishReasonSafety
	default:
		return genai.FinishReasonUnspecified
	}
}
