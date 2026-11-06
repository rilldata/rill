package ai

import (
	"testing"

	"github.com/rilldata/rill/runtime/parser"
	"github.com/stretchr/testify/require"
)

func TestMCPInstructions(t *testing.T) {
	require.Contains(t, MCPInstructions, "## Workflow Overview")
	require.Contains(t, MCPInstructions, "## Skills")
	require.Contains(t, MCPInstructions, "## Project Development")
}

func TestSkillsForAgent(t *testing.T) {
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

	require.Equal(t, []string{"rca", "glossary"}, names(skillsForAgent(skills, parser.SkillAgentAnalyst)))
	require.Equal(t, []string{"glossary", "modeling"}, names(skillsForAgent(skills, parser.SkillAgentDeveloper)))
}
