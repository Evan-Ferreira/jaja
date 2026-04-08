package services

import (
	"context"
	"fmt"

	"os"

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
	Role    AnthropicMessageRole
	Message string
}


type ClaudeServiceConfig struct {
	Model     anthropic.Model
	MaxTokens int64
	Messages  []AnthropicMessage
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
		return nil, fmt.Errorf(err.Error())
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
	}

	if config.Tools != nil {
		params.Tools = *config.Tools
	}
	if config.Betas != nil {
		params.Betas = *config.Betas
	}

	response, err := claudeService.Client.Beta.Messages.New(ctx, params)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Error getting Claude response: %s", err.Error()))
	}

	return response, nil
}

func parseMessagesParam(messages []AnthropicMessage) ([]anthropic.BetaMessageParam, error) {
	res := make([]anthropic.BetaMessageParam, 0, len(messages))

	for _, m := range messages {
		switch m.Role {
		case AnthropicRoleUser:
			res = append(res, anthropic.BetaMessageParam{
				Role:    anthropic.BetaMessageParamRoleUser,
				Content: []anthropic.BetaContentBlockParamUnion{anthropic.NewBetaTextBlock(m.Message)},
			})
		case AnthropicRoleAssistant:
			res = append(res, anthropic.BetaMessageParam{
				Role:    anthropic.BetaMessageParamRoleAssistant,
				Content: []anthropic.BetaContentBlockParamUnion{anthropic.NewBetaTextBlock(m.Message)},
			})
		default:
			return nil, fmt.Errorf("unknown message role: %q", m.Role)
		}
	}

	return res, nil
}