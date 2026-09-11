package ai

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/parser"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMCPInstructions(t *testing.T) {
	require.Contains(t, MCPInstructions, "## Skills")
	require.NotContains(t, mcpInstructionsWithoutSkills, "## Skills")
	// The surrounding sections are present either way
	for _, instr := range []string{MCPInstructions, mcpInstructionsWithoutSkills} {
		require.Contains(t, instr, "## Workflow Overview")
		require.Contains(t, instr, "## Project Development")
	}
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

	// Metrics view names are matched case-insensitively, like the catalog
	require.Equal(t, []string{"rca", "glossary"}, names(filterSkills(skills, parser.SkillAgentAnalyst, []string{"Orders"})))

	// Non-matching metrics view context: scoped skills are excluded, unscoped ones remain
	require.Equal(t, []string{"glossary"}, names(filterSkills(skills, parser.SkillAgentAnalyst, []string{"bids"})))

	// Developer agent
	require.Equal(t, []string{"glossary", "modeling"}, names(filterSkills(skills, parser.SkillAgentDeveloper, nil)))
}

func TestSkillPromptsCap(t *testing.T) {
	// The cap covers the rendered section, so a body that fits on its own but not with its heading falls back to the index.
	skills := []*Skill{
		{Name: "big", Description: "Big skill.", Body: strings.Repeat("x", skillsMaxAlwaysApplyBytes-len("## Skill: big\n\n")), AlwaysApply: true},
		{Name: "small", Description: "Small skill.", Body: "Short.", AlwaysApply: true},
	}
	alwaysApply, index := skillPrompts(skills, zap.NewNop())
	require.Equal(t, "## Skill: small\n\nShort.", alwaysApply)
	require.Equal(t, "- big: Big skill.", index)

	// Many empty bodies still count towards the cap through their headings
	var many []*Skill
	for i := 0; i < skillsMaxAlwaysApplyBytes/10; i++ {
		many = append(many, &Skill{Name: "e", Description: "Empty.", AlwaysApply: true})
	}
	alwaysApply, _ = skillPrompts(many, zap.NewNop())
	require.LessOrEqual(t, len(alwaysApply), skillsMaxAlwaysApplyBytes)

	// The on-demand index is bounded too, and points to list_skills for the rest
	var wide []*Skill
	for i := 0; i < 100; i++ {
		wide = append(wide, &Skill{Name: fmt.Sprintf("s%d", i), Description: strings.Repeat("d", 1024)})
	}
	_, index = skillPrompts(wide, zap.NewNop())
	require.LessOrEqual(t, len(index), skillsMaxIndexBytes+128)
	require.Contains(t, index, "more skills not listed here; call list_skills")
	require.Contains(t, index, "- s0: ")
	require.NotContains(t, index, "- s99: ")
}
