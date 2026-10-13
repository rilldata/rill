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
	memories := []*drivers.AIMemory{
		{ID: "6518299b-ce6c-4c4c-9bb0-c1be022cab86"},
		{ID: "6518299b-ffff-4c4c-9bb0-c1be022cab86"},
		{ID: "cd59e0a1-7ec9-4e8d-a003-2da7bc144318"},
	}

	// The full ID and the prefix shown to the model both resolve
	m, ok := findMemoryByID(memories, memories[2].ID)
	require.True(t, ok)
	require.Equal(t, memories[2], m)
	m, ok = findMemoryByID(memories, shortMemoryID(memories[2].ID))
	require.True(t, ok)
	require.Equal(t, memories[2], m)

	// A longer prefix disambiguates memories that share the shown prefix
	m, ok = findMemoryByID(memories, "6518299b-ffff")
	require.True(t, ok)
	require.Equal(t, memories[1], m)

	// Ambiguous, too short, mistyped and unknown IDs find nothing
	for _, id := range []string{"6518299b", "cd59", "6518299b-ce6c-4c6c", "cd59e0a1-7ec9-4e8d-a003-2da7bc144319", ""} {
		_, ok = findMemoryByID(memories, id)
		require.False(t, ok, id)
	}
}

func TestRenderUserMemories(t *testing.T) {
	require.Empty(t, renderUserMemories(nil, AnalystAgentName))

	// A UUID is shown by its short prefix
	require.Equal(t, "- Prefers tables (id: 6518299b)", renderUserMemories([]*drivers.AIMemory{{ID: "6518299b-ce6c-4c4c-9bb0-c1be022cab86", Content: "Prefers tables"}}, AnalystAgentName))

	// Each agent receives the memories scoped to it plus the ones without an agent
	memories := []*drivers.AIMemory{
		{ID: "a", Agent: AnalystAgentName, Content: "Prefers tables"},
		{ID: "b", Agent: DeveloperAgentName, Content: "Names models in snake_case"},
		{ID: "c", Content: "Works on the growth team"},
	}
	require.Equal(t, "- Prefers tables (id: a)\n- Works on the growth team (id: c)", renderUserMemories(memories, AnalystAgentName))
	require.Equal(t, "- Names models in snake_case (id: b)\n- Works on the growth team (id: c)", renderUserMemories(memories, DeveloperAgentName))

	// The size cap holds
	big := make([]*drivers.AIMemory, 0, MaxMemories)
	for range MaxMemories {
		big = append(big, &drivers.AIMemory{ID: "id", Content: strings.Repeat("y", MaxMemoryContentChars)})
	}
	require.LessOrEqual(t, len(renderUserMemories(big, AnalystAgentName)), maxMemoryPromptChars)
}
