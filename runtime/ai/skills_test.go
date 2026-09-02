package ai_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestSkills(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"skills/revenue-rca.md": `---
description: Playbook for diagnosing revenue drops.
metrics_views: [orders]
---

# Revenue RCA playbook
Always break revenue down by country first.`,
			"skills/glossary/SKILL.md": `---
description: Business glossary.
always_apply: true
agents: [analyst, developer]
---

ARPU excludes trial users.`,
			// Auxiliary file in a skill directory: not a skill.
			"skills/glossary/notes.md": "Internal notes, not a skill.",
			// Missing required description.
			"skills/broken.md": "# No front matter here",
			// Duplicate name (sorted after revenue-rca.md, so it loses).
			"skills/zzz-dupe.md": `---
name: revenue-rca
description: Duplicate of the RCA skill.
---

Body.`,
			// Exceeds the per-file size cap.
			"skills/oversized.md": "---\ndescription: Too big.\n---\n\n" + strings.Repeat("x", 1<<17),
		},
	})
	s := newSession(t, rt, instanceID)

	// List skills: valid skills in path order, invalid files reported with errors
	var listRes *ai.ListSkillsResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListSkillsName, &listRes, &ai.ListSkillsArgs{})
	require.NoError(t, err)
	require.Len(t, listRes.Skills, 2)
	require.Equal(t, "glossary", listRes.Skills[0].Name)
	require.True(t, listRes.Skills[0].AlwaysApply)
	require.Equal(t, []string{"analyst", "developer"}, listRes.Skills[0].Agents)
	require.Equal(t, "revenue-rca", listRes.Skills[1].Name)
	require.Equal(t, []string{"orders"}, listRes.Skills[1].MetricsViews)
	require.Equal(t, []string{"analyst"}, listRes.Skills[1].Agents)

	require.Len(t, listRes.Invalid, 3)
	issues := map[string]string{}
	for _, issue := range listRes.Invalid {
		issues[issue.Path] = issue.Error
	}
	require.Contains(t, issues["/skills/broken.md"], "description")
	require.Contains(t, issues["/skills/oversized.md"], "maximum skill size")
	require.Contains(t, issues["/skills/zzz-dupe.md"], "duplicate skill name")

	// Load a skill by name
	var loadRes *ai.LoadSkillResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.LoadSkillName, &loadRes, &ai.LoadSkillArgs{Name: "revenue-rca"})
	require.NoError(t, err)
	require.Equal(t, "Playbook for diagnosing revenue drops.", loadRes.Description)
	require.Contains(t, loadRes.Body, "Always break revenue down by country first.")
	require.NotContains(t, loadRes.Body, "---")

	// Load an unknown skill: the error lists the available names
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.LoadSkillName, &loadRes, &ai.LoadSkillArgs{Name: "nope"})
	require.ErrorContains(t, err, "glossary, revenue-rca")
}

func TestSkillsEmptyProject(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	s := newSession(t, rt, instanceID)

	var listRes *ai.ListSkillsResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListSkillsName, &listRes, &ai.ListSkillsArgs{})
	require.NoError(t, err)
	require.Empty(t, listRes.Skills)
	require.Empty(t, listRes.Invalid)

	var loadRes *ai.LoadSkillResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.LoadSkillName, &loadRes, &ai.LoadSkillArgs{Name: "anything"})
	require.ErrorContains(t, err, "does not define any skills")
}

// TestSkillsInListMetricsViews verifies that always-apply skills are appended to the
// ai_instructions field of list_metrics_views for external MCP clients,
// but not for Rill's own agents (which receive them directly in their prompts).
func TestSkillsInListMetricsViews(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"skills/glossary.md": `---
description: Business glossary.
always_apply: true
---

ARPU excludes trial users.`,
			"skills/on-demand.md": `---
description: An on-demand skill.
---

Not injected wholesale.`,
		},
	})

	newSessionWithUserAgent := func(t *testing.T, userAgent string) *ai.Session {
		claims := &runtime.SecurityClaims{UserID: uuid.NewString(), SkipChecks: true}
		r := ai.NewRunner(rt, activity.NewNoopClient())
		s, err := r.Session(t.Context(), &ai.SessionOptions{
			InstanceID: instanceID,
			Claims:     claims,
			UserAgent:  userAgent,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, s.Flush(t.Context()))
		})
		return s
	}

	// External MCP client: always-apply skill body is included, on-demand skill is not
	s := newSessionWithUserAgent(t, "mcp-client")
	var res *ai.ListMetricsViewsResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListMetricsViewsName, &res, &ai.ListMetricsViewsArgs{})
	require.NoError(t, err)
	require.Contains(t, res.AIInstructions, "## Skill: glossary")
	require.Contains(t, res.AIInstructions, "ARPU excludes trial users.")
	require.NotContains(t, res.AIInstructions, "Not injected wholesale.")

	// Rill's own agents: no enrichment (skills are injected into their prompts instead)
	s = newSessionWithUserAgent(t, "rill-web")
	res = nil
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.ListMetricsViewsName, &res, &ai.ListMetricsViewsArgs{})
	require.NoError(t, err)
	require.Empty(t, res.AIInstructions)
}

// TestSkillsMCPAccess verifies that the skill tools are exposed to any principal with UseAI,
// including viewers without repo access.
func TestSkillsMCPAccess(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})

	newMCPSession := func(t *testing.T, permissions ...runtime.Permission) *ai.Session {
		claims := &runtime.SecurityClaims{
			UserID:      uuid.NewString(),
			SkipChecks:  false,
			Permissions: permissions,
		}
		r := ai.NewRunner(rt, activity.NewNoopClient())
		s, err := r.Session(t.Context(), &ai.SessionOptions{
			InstanceID: instanceID,
			Claims:     claims,
			UserAgent:  "mcp-client",
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, s.Flush(t.Context()))
		})
		return s
	}

	assertAccess := func(t *testing.T, s *ai.Session, name string, want bool) {
		t.Helper()
		tool, ok := s.Tool(name)
		require.True(t, ok, "tool %q should be registered", name)
		allowed, err := tool.CheckAccess(ai.WithSession(t.Context(), s))
		require.NoError(t, err)
		require.Equal(t, want, allowed, "tool %q access", name)
	}

	// A viewer-like claim set (no ReadRepo/EditRepo) can use the skill tools
	s := newMCPSession(t, runtime.UseAI, runtime.ReadMetrics, runtime.ReadObjects)
	assertAccess(t, s, ai.ListSkillsName, true)
	assertAccess(t, s, ai.LoadSkillName, true)

	// Without UseAI, the skill tools are not accessible
	s = newMCPSession(t, runtime.ReadMetrics, runtime.ReadObjects)
	assertAccess(t, s, ai.ListSkillsName, false)
	assertAccess(t, s, ai.LoadSkillName, false)
}
