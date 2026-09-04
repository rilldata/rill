package parser

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestSkill(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		// Valid skill with Rill extension fields
		`skills/revenue-rca/SKILL.md`: `---
name: revenue-rca
description: Playbook for diagnosing revenue drops.
metrics_views: [orders]
agents: [analyst, developer]
---

# Revenue RCA playbook
Always break revenue down by country first.`,
		// Valid skill in the generic .agents/skills directory, with only the standard fields
		`.agents/skills/glossary/SKILL.md`: `---
name: glossary
description: Business glossary.
license: Apache-2.0
metadata:
  author: example-org
---

ARPU excludes trial users.`,
		// Name defaults to the directory name; always_apply is a Rill extension
		`skills/formatting/SKILL.md`: `---
description: House formatting rules.
always_apply: true
---

Report percentages with one decimal.`,
		// Markdown files that are not skills are ignored
		`skills/revenue-rca/references/notes.md`: `Auxiliary file, not a skill.`,
		`skills/loose.md`:                        `Not in a skill directory.`,
		`README.md`:                              `Not a skill.`,
		// Invalid: missing description
		`skills/broken/SKILL.md`: `---
name: broken
---

Body.`,
		// Invalid: empty front matter, which must be reported as the missing description rather than as an unclosed delimiter
		`skills/empty/SKILL.md`: `---
---

Body.`,
		// Invalid: front matter name doesn't match the directory name
		`skills/mismatch/SKILL.md`: `---
name: other-name
description: Mismatched name.
---

Body.`,
		// Invalid: unknown front matter field (probably a typo)
		`skills/typo/SKILL.md`: `---
description: Unknown field.
metric_views: [orders]
---

Body.`,
		// Invalid: directory name violates the naming rules
		`skills/Bad_Name/SKILL.md`: `---
description: Invalid name.
---

Body.`,
	})

	resources := []*Resource{
		{
			Name:  ResourceName{Kind: ResourceKindSkill, Name: "revenue-rca"},
			Paths: []string{"/skills/revenue-rca/SKILL.md"},
			Refs:  []ResourceName{{Kind: ResourceKindMetricsView, Name: "orders"}},
			SkillSpec: &runtimev1.SkillSpec{
				Description:  "Playbook for diagnosing revenue drops.",
				Body:         "# Revenue RCA playbook\nAlways break revenue down by country first.",
				MetricsViews: []string{"orders"},
				Agents:       []string{"analyst", "developer"},
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindSkill, Name: "glossary"},
			Paths: []string{"/.agents/skills/glossary/SKILL.md"},
			SkillSpec: &runtimev1.SkillSpec{
				Description: "Business glossary.",
				Body:        "ARPU excludes trial users.",
				Agents:      []string{"analyst"},
			},
		},
		{
			Name:  ResourceName{Kind: ResourceKindSkill, Name: "formatting"},
			Paths: []string{"/skills/formatting/SKILL.md"},
			SkillSpec: &runtimev1.SkillSpec{
				Description: "House formatting rules.",
				Body:        "Report percentages with one decimal.",
				Agents:      []string{"analyst"},
				AlwaysApply: true,
			},
		},
	}
	perrors := []*runtimev1.ParseError{
		{FilePath: "/skills/broken/SKILL.md", Message: `missing required front matter field "description"`},
		{FilePath: "/skills/empty/SKILL.md", Message: `missing required front matter field "description"`},
		{FilePath: "/skills/mismatch/SKILL.md", Message: `must match the skill's directory name`},
		{FilePath: "/skills/typo/SKILL.md", Message: "failed to parse front matter"},
		{FilePath: "/skills/Bad_Name/SKILL.md", Message: "invalid skill name"},
	}

	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	requireResourcesAndErrors(t, p, resources, perrors)

	// Incremental reparse: edit a skill file
	putRepo(t, repo, map[string]string{
		`skills/formatting/SKILL.md`: `---
description: House formatting rules.
always_apply: true
---

Report percentages with two decimals.`,
	})
	diff, err := p.Reparse(ctx, []string{"/skills/formatting/SKILL.md"})
	require.NoError(t, err)
	require.Equal(t, []ResourceName{{Kind: ResourceKindSkill, Name: "formatting"}}, diff.Modified)
	require.Equal(t, "Report percentages with two decimals.", p.Resources[ResourceName{Kind: ResourceKindSkill, Name: "formatting"}.Normalized()].SkillSpec.Body)

	// Incremental reparse: delete a skill file
	deleteRepo(t, repo, "/skills/formatting/SKILL.md")
	diff, err = p.Reparse(ctx, []string{"/skills/formatting/SKILL.md"})
	require.NoError(t, err)
	require.Equal(t, []ResourceName{{Kind: ResourceKindSkill, Name: "formatting"}}, diff.Deleted)
}

func TestSkillNameForPath(t *testing.T) {
	tests := []struct {
		path     string
		wantName string
		wantOK   bool
	}{
		{"/skills/revenue-rca/SKILL.md", "revenue-rca", true},
		{"/.agents/skills/glossary/SKILL.md", "glossary", true},
		{"/skills/loose.md", "", false},
		{"/skills/glossary/notes.md", "", false},
		{"/skills/a/b/SKILL.md", "", false},
		{"/models/orders.md", "", false},
		{"/SKILL.md", "", false},
	}
	for _, tt := range tests {
		name, ok := skillNameForPath(tt.path)
		require.Equal(t, tt.wantOK, ok, "path %q", tt.path)
		require.Equal(t, tt.wantName, name, "path %q", tt.path)
	}
}
