package gemini

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

func TestUsageTokens(t *testing.T) {
	tests := []struct {
		name       string
		usage      *genai.GenerateContentResponseUsageMetadata
		wantInput  int
		wantCached int
		wantOutput int
	}{
		{
			name:       "without thinking",
			usage:      &genai.GenerateContentResponseUsageMetadata{PromptTokenCount: 1000, CandidatesTokenCount: 200},
			wantInput:  1000,
			wantOutput: 200,
		},
		{
			name:       "thinking tokens are billed as output",
			usage:      &genai.GenerateContentResponseUsageMetadata{PromptTokenCount: 1000, CandidatesTokenCount: 200, ThoughtsTokenCount: 300},
			wantInput:  1000,
			wantOutput: 500,
		},
		{
			name:       "cached tokens are a subset of input",
			usage:      &genai.GenerateContentResponseUsageMetadata{PromptTokenCount: 1000, CachedContentTokenCount: 800, CandidatesTokenCount: 200, ThoughtsTokenCount: 300},
			wantInput:  1000,
			wantCached: 800,
			wantOutput: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, cached, output := usageTokens(tt.usage)
			require.Equal(t, tt.wantInput, input)
			require.Equal(t, tt.wantCached, cached)
			require.Equal(t, tt.wantOutput, output)
		})
	}
}
