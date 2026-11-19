package ai

import (
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestValidateMemoryContent(t *testing.T) {
	cases := []struct {
		content string
		ok      bool
	}{
		{"Prefers week-over-week comparisons", true},
		{"When they say core markets they mean US and UK", true},
		{"Compares 2024-01-01 2025-01-01 when asked about growth", true},
		{"Uses fiscal year 2025 with 4 quarters and 13 periods", true},
		{"Ignores test accounts when counting active users", true},
		{"Wants the row limit raised to 500 in tables", true},
		{"Looks at the risk-register dashboard first", true}, // "sk-" inside a word is not a key
		{"Analyzes token spend per model", true},
		{"", false},
		{strings.Repeat("x", MaxMemoryContentChars+1), false},
		{"The api key is abc123", false},
		{"Their password is hunter2", false},
		{"Uses sk-abc123def456ghi789 for the API", false},
		{"GitHub token is ghp_abcdefghijklmnop1234", false},
		{"AWS key AKIAIOSFODNN7EXAMPLE", false},
		{"Slack bot xoxb-1234567890-abcdefghij", false},
		{"Contact them at jane@example.com", false},
	}
	for _, c := range cases {
		reason := validateMemoryContent(c.content)
		if c.ok {
			require.Empty(t, reason, c.content)
		} else {
			require.NotEmpty(t, reason, c.content)
		}
	}
}

func TestSanitizeMemoryContent(t *testing.T) {
	require.Equal(t, "Prefers tables over charts", SanitizeMemoryContent("  Prefers\ttables\n over  charts "))
	require.Equal(t, "Prefers charts", SanitizeMemoryContent("Prefers <charts>"))
	require.Equal(t, "/user_memory", SanitizeMemoryContent("</user_memory>"))
	require.Empty(t, SanitizeMemoryContent(" <> "))
}

func TestNormalizeMemoryContent(t *testing.T) {
	require.Equal(t, normalizeMemoryContent("Prefers  Tables."), normalizeMemoryContent("prefers tables"))
	require.NotEqual(t, normalizeMemoryContent("Prefers tables"), normalizeMemoryContent("Prefers charts"))
}

func TestFindMemoryByID(t *testing.T) {
	memories := []*drivers.AIMemory{{ID: 1}, {ID: 2}, {ID: 10}}

	m, ok := findMemoryByID(memories, 10)
	require.True(t, ok)
	require.Equal(t, memories[2], m)

	for _, id := range []int64{0, 3, 100} {
		_, ok = findMemoryByID(memories, id)
		require.False(t, ok, id)
	}
}

func TestRenderUserMemories(t *testing.T) {
	require.Empty(t, renderUserMemories(nil, AnalystAgentName))

	// Each agent receives the memories scoped to it plus the ones without an agent
	memories := []*drivers.AIMemory{
		{ID: 1, Agent: AnalystAgentName, Content: "Prefers tables"},
		{ID: 2, Agent: DeveloperAgentName, Content: "Names models in snake_case"},
		{ID: 3, Content: "Works on the growth team"},
	}
	require.Equal(t, "<memory id=\"1\">Prefers tables</memory>\n<memory id=\"3\">Works on the growth team</memory>", renderUserMemories(memories, AnalystAgentName))
	require.Equal(t, "<memory id=\"2\">Names models in snake_case</memory>\n<memory id=\"3\">Works on the growth team</memory>", renderUserMemories(memories, DeveloperAgentName))

	// The size cap holds
	big := make([]*drivers.AIMemory, 0, MaxMemories)
	for i := range MaxMemories {
		big = append(big, &drivers.AIMemory{ID: int64(i + 1), Content: strings.Repeat("y", MaxMemoryContentChars)})
	}
	require.LessOrEqual(t, len(renderUserMemories(big, AnalystAgentName)), maxMemoryPromptChars)
}
