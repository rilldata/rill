package ai

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestPromptToTitle_multibyteTruncationStaysValidUTF8(t *testing.T) {
	// Live Rill_UO prompt whose previous byte-wise cut (title[:47]+"...") split 有.
	prompt := "今年乘用车增长快于大盘的区域都有哪些"
	require.Greater(t, len(prompt), 50)

	oldCut := prompt[:47] + "..."
	require.False(t, utf8.ValidString(oldCut), "the previous cut point must be invalid UTF-8 (the bug)")

	title := promptToTitle(prompt)
	require.True(t, utf8.ValidString(title))
	require.LessOrEqual(t, len(title), 50)
	require.True(t, strings.HasSuffix(title, "..."))
	require.Equal(t, "今年乘用车增长快于大盘的区域都...", title)
}

func TestPromptToTitle_asciiStillTruncatesAt50Bytes(t *testing.T) {
	title := promptToTitle(strings.Repeat("a", 80))
	require.Equal(t, strings.Repeat("a", 47)+"...", title)
	require.True(t, utf8.ValidString(title))
}

func TestPromptToTitle_emptyFallsBack(t *testing.T) {
	require.Equal(t, "New Conversation", promptToTitle("   \n\t  "))
}
