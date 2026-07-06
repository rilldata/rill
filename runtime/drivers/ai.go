package drivers

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
)

// DefaultAIMaxInputTokens is a conservative default input token limit for AI completion requests.
// It matches the smallest context window among the commonly used OpenAI, Gemini, and Claude models
// (Claude's 200k; GPT-5.x allows 272k and Gemini ~1M). Drivers can override it via their
// max_input_tokens config property.
const DefaultAIMaxInputTokens = 200_000

type AIService interface {
	Complete(ctx context.Context, opts *CompleteOptions) (*CompleteResult, error)
	// MaxInputTokens returns the maximum number of input tokens a completion request may contain.
	// It is configurable per connector via the max_input_tokens property and defaults to DefaultAIMaxInputTokens.
	MaxInputTokens() int
}

type CompleteOptions struct {
	Messages     []*aiv1.CompletionMessage
	Tools        []*aiv1.Tool
	OutputSchema *jsonschema.Schema
	// CacheKey identifies a series of requests that share a prompt prefix (e.g. an AI session ID).
	// Providers may use it to improve prompt cache routing. Drivers that don't support it ignore it.
	CacheKey string
}

type CompleteResult struct {
	Message *aiv1.CompletionMessage
	// Provider is the LLM provider that served the completion (e.g. "claude", "openai", "gemini"). For the managed admin
	// proxy this is the real underlying provider, not "admin", so it stays meaningful through the proxy.
	Provider string
	// InputTokens and CachedInputTokens are reported as the provider returns them, so their relationship is
	// provider-specific: for Claude they are disjoint (InputTokens excludes cached); for OpenAI/Gemini CachedInputTokens
	// is a subset of InputTokens.
	InputTokens       int
	CachedInputTokens int
	OutputTokens      int
}
