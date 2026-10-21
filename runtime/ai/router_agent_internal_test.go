package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromptToTitle(t *testing.T) {
	mvRef := `<chat-reference>type="metricsView" metricsView="orders"</chat-reference>`

	cases := []struct {
		name   string
		prompt string
		title  string
	}{
		{
			name:   "empty",
			prompt: " \n ",
			title:  "New Conversation",
		},
		{
			name:   "whitespace is collapsed",
			prompt: "  How did\n\nrevenue  change? ",
			title:  "How did revenue change?",
		},
		{
			name:   "long prompt is truncated",
			prompt: strings.Repeat("a", 60),
			title:  strings.Repeat("a", 47) + "...",
		},
		{
			name:   "length is counted in runes",
			prompt: strings.Repeat("ú", 50),
			title:  strings.Repeat("ú", 50),
		},
		{
			name:   "truncation doesn't split a rune",
			prompt: strings.Repeat("ú", 51),
			title:  strings.Repeat("ú", 47) + "...",
		},
		{
			name:   "reference counts as 10 runes",
			prompt: mvRef + " " + strings.Repeat("a", 39),
			title:  mvRef + " " + strings.Repeat("a", 39),
		},
		{
			name:   "text after a reference is truncated",
			prompt: mvRef + " " + strings.Repeat("a", 50),
			title:  mvRef + " " + strings.Repeat("a", 36) + "...",
		},
		{
			name:   "reference that doesn't fit is cut before",
			prompt: strings.Repeat("a", 40) + " " + mvRef + " by month",
			title:  strings.Repeat("a", 40) + "...",
		},
		{
			name:   "reference that fits exactly is kept",
			prompt: strings.Repeat("a", 37) + mvRef + " by month",
			title:  strings.Repeat("a", 37) + mvRef + "...",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.title, promptToTitle(c.prompt))
		})
	}
}
