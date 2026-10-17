package ai

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestPromptToTitle(t *testing.T) {
	cases := []struct {
		name   string
		prompt string
		want   string
	}{
		{
			name:   "short ascii is unchanged",
			prompt: "show sessions",
			want:   "show sessions",
		},
		{
			name:   "whitespace is collapsed and trimmed",
			prompt: "  show   me \n sessions ",
			want:   "show me sessions",
		},
		{
			name:   "empty falls back",
			prompt: "   ",
			want:   "New Conversation",
		},
		{
			name:   "long ascii is truncated to 50 runes",
			prompt: strings.Repeat("a", 100),
			want:   strings.Repeat("a", 47) + "...",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, promptToTitle(c.prompt))
		})
	}
}

// TestPromptToTitleMultibyteStaysValid guards the UTF-8 bug: truncating a title
// by byte offset can cut through the middle of a multi-byte rune (e.g. CJK),
// producing invalid UTF-8. The title must always remain valid UTF-8.
func TestPromptToTitleMultibyteStaysValid(t *testing.T) {
	// 60 Korean syllables (3 bytes each = 180 bytes); a byte-offset cut at 47
	// would land mid-rune.
	long := strings.Repeat("가", 60)
	got := promptToTitle(long)

	require.True(t, utf8.ValidString(got), "title must stay valid UTF-8: %q", got)
	// 47 kept runes + the "..." suffix.
	require.Equal(t, strings.Repeat("가", 47)+"...", got)
	require.Equal(t, 50, utf8.RuneCountInString(got))
}
