package agent

import (
	"context"
	"fmt"

	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)
type Agent struct {
	client *anthropic.Client
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



func New() (*Agent, error){
	client, err := newClient()

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return &Agent{
		client,
	}, nil
}

func (agent *Agent) Run(ctx context.Context, model anthropic.Model, prompt string, fileUrl *string) (*anthropic.Message, error){
	message := anthropic.NewUserMessage(
		anthropic.NewDocumentBlock(anthropic.URLPDFSourceParam{
			URL: *fileUrl,
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

