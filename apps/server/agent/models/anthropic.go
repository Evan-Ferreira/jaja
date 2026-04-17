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
	"iter"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/joho/godotenv"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

const defaultMaxTokens int64 = 8192

var validAnthropicModels = map[anthropic.Model]struct{}{
	anthropic.ModelClaudeOpus4_6:            {},
	anthropic.ModelClaudeSonnet4_6:          {},
	anthropic.ModelClaudeHaiku4_5:           {},
	anthropic.ModelClaudeHaiku4_5_20251001:  {},
	anthropic.ModelClaudeOpus4_5:            {},
	anthropic.ModelClaudeOpus4_5_20251101:   {},
	anthropic.ModelClaudeSonnet4_5:          {},
	anthropic.ModelClaudeSonnet4_5_20250929: {},
	anthropic.ModelClaudeOpus4_1:            {},
	anthropic.ModelClaudeOpus4_1_20250805:   {},
	anthropic.ModelClaudeOpus4_0:            {},
	anthropic.ModelClaudeOpus4_20250514:     {},
	anthropic.ModelClaudeSonnet4_0:          {},
	anthropic.ModelClaudeSonnet4_20250514:   {},
}

// AnthropicModel implements model.LLM using the official Anthropic Go SDK.
type AnthropicModel struct {
	client    *anthropic.Client
	modelName anthropic.Model
	maxTokens int64
}

// NewAnthropicModel builds an AnthropicModel using ANTHROPIC_API_KEY from env.
func NewAnthropicModel(modelName anthropic.Model) (*AnthropicModel, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	if _, ok := validAnthropicModels[modelName]; !ok {
		return nil, fmt.Errorf("unknown Anthropic model %q: see anthropic.Model constants for valid values", modelName)
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return NewAnthropicModelWithClient(modelName, &client), nil
}

// NewAnthropicModelWithClient is the Go analogue of Java's
// `new Claude(modelName, anthropicClient)` — it takes a caller-constructed
// client so callers can customize transport, base URL, auth, etc.
func NewAnthropicModelWithClient(modelName anthropic.Model, client *anthropic.Client) *AnthropicModel {
	return &AnthropicModel{
		client:    client,
		modelName: modelName,
		maxTokens: defaultMaxTokens,
	}
}

// generateStream streams responses from the Anthropic API, yielding text deltas
// as partial responses and a final complete response with TurnComplete set to true.
func (m *AnthropicModel) generateStream(ctx context.Context, req *model.LLMRequest) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		params, err := m.buildParams(req)
		if err != nil {
			yield(nil, fmt.Errorf("failed to build Anthropic params: %w", err))
			return
		}

		stream := m.client.Messages.NewStreaming(ctx, params)
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

// WithMaxTokens overrides the default max output tokens (8192).
func (m *AnthropicModel) WithMaxTokens(n int64) *AnthropicModel {
	m.maxTokens = n
	return m
}

func (m *AnthropicModel) Name() string { return m.modelName }

// GenerateContent implements model.LLM. The `stream` argument is ignored —
// streaming is not yet implemented for Claude, matching adk-java behavior.
func (m *AnthropicModel) GenerateContent(ctx context.Context, req *model.LLMRequest, _ bool, stream bool) iter.Seq2[*model.LLMResponse, error] {
	if stream {
		return m.generateStream(ctx, req)
	}
	return func(yield func(*model.LLMResponse, error) bool) {
		params, err := m.buildParams(req)
		if err != nil {
			yield(nil, fmt.Errorf("build anthropic params: %w", err))
			return
		}

		msg, err := m.client.Messages.New(ctx, params)
		if err != nil {
			yield(nil, fmt.Errorf("anthropic messages.new: %w", err))
			return
		}
		yield(convertResponse(msg), nil)
	}
}

func (m *AnthropicModel) buildParams(req *model.LLMRequest) (anthropic.MessageNewParams, error) {
	modelName := m.modelName
	if req.Model != "" {
		modelName = req.Model
	}

	params := anthropic.MessageNewParams{
		Model:     modelName,
		MaxTokens: m.maxTokens,
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
			blocks = append(blocks, anthropic.NewTextBlock(
				fmt.Sprintf("[file: %s (%s)]", p.FileData.FileURI, p.FileData.MIMEType),
			))
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
				InputSchema: buildInputSchema(fd.Parameters),
			}
			if fd.Description != "" {
				tp.Description = param.NewOpt(fd.Description)
			}
			out = append(out, anthropic.ToolUnionParam{OfTool: &tp})
		}
	}
	return out
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
