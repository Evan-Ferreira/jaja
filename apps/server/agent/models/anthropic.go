package models

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// set to check valid model names
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

type AnthropicModel struct {
	client    *anthropic.Client
	modelName string
}

func NewAnthropicModel(modelName anthropic.Model) (*AnthropicModel, error) {
	if _, ok := validAnthropicModels[modelName]; !ok {
		return nil, fmt.Errorf("Unknown Anthropic model %q: see anthropic.Model constants for valid values", modelName)
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	return &AnthropicModel{
		client:    &client,
		modelName: modelName,
	}, nil
}

func (m *AnthropicModel) Name() string {
	return m.modelName
}

func (m *AnthropicModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	if stream {
		return m.generateStream(ctx, req)
	}

	return func(yield func(*model.LLMResponse, error) bool) {
		params, err := buildParams(m.modelName, req)
		if err != nil {
			yield(nil, fmt.Errorf("failed to build Anthropic params: %w", err))
			return
		}

		resp, err := m.client.Messages.New(ctx, params)
		if err != nil {
			yield(nil, fmt.Errorf("Anthropic API call failed: %w", err))
			return
		}

		yield(convertResponse(resp), nil)
	}
}

// generateStream streams responses from the Anthropic API, yielding text deltas
// as partial responses and a final complete response with TurnComplete set to true.
func (m *AnthropicModel) generateStream(ctx context.Context, req *model.LLMRequest) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		params, err := buildParams(m.modelName, req)
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

// buildParams converts an ADK LLMRequest into Anthropic MessageNewParams.
func buildParams(modelName string, req *model.LLMRequest) (anthropic.MessageNewParams, error) {
	params := anthropic.MessageNewParams{
		Model:     modelName,
		MaxTokens: 4096,
	}

	if req.Config != nil && req.Config.SystemInstruction != nil {
		params.System = convertSystemInstruction(req.Config.SystemInstruction)
	}

	msgs, err := convertMessages(req.Contents)
	if err != nil {
		return params, err
	}
	params.Messages = msgs

	if req.Config != nil {
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
	}

	return params, nil
}

// convertSystemInstruction extracts text parts from a genai.Content into Anthropic system blocks.
func convertSystemInstruction(content *genai.Content) []anthropic.TextBlockParam {
	var blocks []anthropic.TextBlockParam
	for _, part := range content.Parts {
		if part.Text != "" {
			blocks = append(blocks, anthropic.TextBlockParam{Text: part.Text})
		}
	}
	return blocks
}

// convertMessages converts ADK genai.Content messages to Anthropic MessageParams.
func convertMessages(contents []*genai.Content) ([]anthropic.MessageParam, error) {
	var msgs []anthropic.MessageParam
	for _, c := range contents {
		if c == nil {
			continue
		}
		blocks, err := convertParts(c.Parts)
		if err != nil {
			return nil, fmt.Errorf("converting parts for role %q: %w", c.Role, err)
		}
		if len(blocks) == 0 {
			continue
		}
		role := convertRole(c.Role)
		msgs = append(msgs, anthropic.MessageParam{Role: role, Content: blocks})
	}
	return msgs, nil
}

// convertRole maps genai roles to Anthropic roles.
func convertRole(role string) anthropic.MessageParamRole {
	if role == "model" {
		return anthropic.MessageParamRoleAssistant
	}
	return anthropic.MessageParamRoleUser
}

// convertParts converts genai Parts to Anthropic content blocks.
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
			id := fc.ID
			if id == "" {
				id = "call_" + fc.Name
			}
			blocks = append(blocks, anthropic.NewToolUseBlock(id, fc.Args, fc.Name))

		case p.FunctionResponse != nil:
			fr := p.FunctionResponse
			id := fr.ID
			if id == "" {
				id = "call_" + fr.Name
			}
			content, err := json.Marshal(fr.Response)
			if err != nil {
				return nil, fmt.Errorf("marshaling function response %q: %w", fr.Name, err)
			}
			blocks = append(blocks, anthropic.NewToolResultBlock(id, string(content), false))

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

// convertResponse converts an Anthropic Message to an ADK LLMResponse.
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

// convertResponseBlocks converts Anthropic content blocks to genai Parts.
func convertResponseBlocks(blocks []anthropic.ContentBlockUnion) []*genai.Part {
	var parts []*genai.Part
	for _, b := range blocks {
		switch b.Type {
		case "text":
			parts = append(parts, &genai.Part{Text: b.Text})
		case "tool_use":
			var args map[string]any
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

// convertStopReason maps Anthropic stop reasons to genai FinishReasons.
// tool_use maps to FinishReasonStop so ADK recognises it as a complete turn
// and proceeds to execute the requested tool before calling back into the model.
func convertStopReason(reason anthropic.StopReason) genai.FinishReason {
	switch reason {
	case anthropic.StopReasonEndTurn, anthropic.StopReasonStopSequence, anthropic.StopReasonToolUse:
		return genai.FinishReasonStop
	case anthropic.StopReasonMaxTokens:
		return genai.FinishReasonMaxTokens
	default:
		return genai.FinishReasonUnspecified
	}
}
