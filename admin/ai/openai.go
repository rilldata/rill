package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

const openAIURL = "https://api.openai.com"

type openAI struct {
	apiKey string
}

var _ Client = (*openAI)(nil)

func NewOpenAI(apiKey string) (Client, error) {
	return &openAI{
		apiKey: apiKey,
	}, nil
}

func (c *openAI) Complete(ctx context.Context, prompt string) (string, error) {
	res, err := c.request(ctx, "/v1/completions", map[string]any{
		"model":       "text-davinci-003",
		"prompt":      prompt,
		"temperature": 0.7,
		"max_tokens":  100,
	})
	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(res)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (c *openAI) request(ctx context.Context, path string, payload map[string]any) (map[string]any, error) {
	uri, err := url.JoinPath(openAIURL, path)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}
