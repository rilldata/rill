package ai

import (
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/parser"
	"github.com/stretchr/testify/require"
)

func TestMCPInstructionsFor(t *testing.T) {
	require.Contains(t, mcpInstructionsFor(true), "## Skills")
	require.NotContains(t, mcpInstructionsFor(false), "## Skills")
	// The surrounding sections are present either way
	for _, hasSkills := range []bool{true, false} {
		instr := mcpInstructionsFor(hasSkills)
		require.Contains(t, instr, "## Workflow Overview")
		require.Contains(t, instr, "## Project Development")
	}
	require.True(t, strings.Contains(MCPInstructions, "## Skills"))
}

func TestFilterSkills(t *testing.T) {
	skills := []*Skill{
		{Name: "rca", Agents: []string{parser.SkillAgentAnalyst}, MetricsViews: []string{"orders"}},
		{Name: "glossary", Agents: []string{parser.SkillAgentAnalyst, parser.SkillAgentDeveloper}},
		{Name: "modeling", Agents: []string{parser.SkillAgentDeveloper}},
	}

	names := func(skills []*Skill) []string {
		res := make([]string, len(skills))
		for i, sk := range skills {
			res[i] = sk.Name
		}
		return res
	}

	// No metrics view context: all analyst skills are included (scoping is relevance, not security)
	require.Equal(t, []string{"rca", "glossary"}, names(filterSkills(skills, parser.SkillAgentAnalyst, nil)))

	// Matching metrics view context
	require.Equal(t, []string{"rca", "glossary"}, names(filterSkills(skills, parser.SkillAgentAnalyst, []string{"orders"})))

	// Non-matching metrics view context: scoped skills are excluded, unscoped ones remain
	require.Equal(t, []string{"glossary"}, names(filterSkills(skills, parser.SkillAgentAnalyst, []string{"bids"})))

	// Developer agent
	require.Equal(t, []string{"glossary", "modeling"}, names(filterSkills(skills, parser.SkillAgentDeveloper, nil)))
}
