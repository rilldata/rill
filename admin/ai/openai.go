package ai

import (
	"context"
	"errors"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/sashabaranov/go-openai"
)

type openAI struct {
	client *openai.Client
	apiKey string
}

var _ Client = (*openAI)(nil)

func NewOpenAI(apiKey string) (Client, error) {
	c := openai.NewClient(apiKey)

	return &openAI{
		client: c,
		apiKey: apiKey,
	}, nil
}

func (c *openAI) Complete(ctx context.Context, msgs []*adminv1.CompletionMessage) (*adminv1.CompletionMessage, error) {
	reqMsgs := make([]openai.ChatCompletionMessage, len(msgs))
	for i, msg := range msgs {
		reqMsgs[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Data,
		}
	}

	res, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Messages:    reqMsgs,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, err
	}

	if len(res.Choices) == 0 {
		return nil, errors.New("no choices returned")
	}

	return &adminv1.CompletionMessage{
		Role: openai.ChatMessageRoleAssistant,
		Data: res.Choices[0].Message.Content,
	}, nil
}
