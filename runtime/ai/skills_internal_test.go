package ai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSkillNameForPath(t *testing.T) {
	tests := []struct {
		path     string
		wantName string
		wantOK   bool
	}{
		{"/skills/revenue-rca.md", "revenue-rca", true},
		{"/skills/glossary/SKILL.md", "glossary", true},
		{"/skills/glossary/notes.md", "", false},
		{"/skills/a/b/SKILL.md", "", false},
		{"/models/orders.md", "", false},
		{"/skills.md", "", false},
	}
	for _, tt := range tests {
		name, ok := skillNameForPath(tt.path)
		require.Equal(t, tt.wantOK, ok, "path %q", tt.path)
		require.Equal(t, tt.wantName, name, "path %q", tt.path)
	}
}

func TestFilterSkills(t *testing.T) {
	skills := []*Skill{
		{Name: "rca", Agents: []string{skillAgentAnalyst}, MetricsViews: []string{"orders"}},
		{Name: "glossary", Agents: []string{skillAgentAnalyst, skillAgentDeveloper}},
		{Name: "modeling", Agents: []string{skillAgentDeveloper}},
	}

	names := func(skills []*Skill) []string {
		res := make([]string, len(skills))
		for i, sk := range skills {
			res[i] = sk.Name
		}
		return res
	}

	// No metrics view context: all analyst skills are included (scoping is relevance, not security)
	require.Equal(t, []string{"rca", "glossary"}, names(filterSkills(skills, skillAgentAnalyst, nil)))

	// Matching metrics view context
	require.Equal(t, []string{"rca", "glossary"}, names(filterSkills(skills, skillAgentAnalyst, []string{"orders"})))

	// Non-matching metrics view context: scoped skills are excluded, unscoped ones remain
	require.Equal(t, []string{"glossary"}, names(filterSkills(skills, skillAgentAnalyst, []string{"bids"})))

	// Developer agent
	require.Equal(t, []string{"glossary", "modeling"}, names(filterSkills(skills, skillAgentDeveloper, nil)))
}
