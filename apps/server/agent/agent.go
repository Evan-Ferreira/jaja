package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)
type Agent struct {
	client *anthropic.Client
}

type AnthropicPDF struct {
	PresignedURL string
}

func New() (*Agent, error){
	client, err := newClient()

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return &Agent{
		client,
	}, nil
}

func (agent *Agent) Run(ctx context.Context, model anthropic.Model, prompt string, anthropicPDF *AnthropicPDF) (*anthropic.Message, error){
	message := anthropic.NewUserMessage(
		anthropic.NewDocumentBlock(anthropic.URLPDFSourceParam{
			URL: anthropicPDF.PresignedURL,
		}),
		anthropic.NewTextBlock(prompt),
	)

	response, err := agent.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: 1024,
		Messages:  []anthropic.MessageParam{message},
	})

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Error getting Claude response: %s", err.Error()))
	}

	return response, nil
}

