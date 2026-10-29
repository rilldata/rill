package ai_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestSkills(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/orders.sql": `SELECT 1 AS revenue`,
			"metrics/orders.yaml": `
type: metrics_view
model: orders
measures:
- name: revenue
  expression: SUM(revenue)
`,
			"skills/revenue-rca/SKILL.md": `---
name: revenue-rca
description: Playbook for diagnosing revenue drops.
metrics_views: [orders]
agents: [analyst]
---

# Revenue RCA playbook
Always break revenue down by country first.`,
			// Skills are also loaded from the generic .agents/skills directory
			".agents/skills/glossary/SKILL.md": `---
description: Business glossary.
always_apply: true
agents: [analyst, developer]
---

ARPU excludes trial users.`,
			// Scoped to a metrics view that doesn't exist: fails reconciliation and must not reach the agents
			"skills/stale/SKILL.md": `---
description: Scoped to a missing metrics view.
metrics_views: [missing]
---

Stale instructions.`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 7, 1, 0)
	s := newSession(t, rt, instanceID)

	// List skills: sorted by name
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

// TestSkillsPreload verifies that preloading seeds a list_skills call and a load_skill call
// for each of the agent's always-apply skills, and nothing else.
func TestSkillsPreload(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"skills/glossary/SKILL.md": `---
description: Business glossary.
agents: [analyst]
always_apply: true
---

ARPU excludes trial users.`,
			"skills/revenue-rca/SKILL.md": `---
description: Revenue playbook.
agents: [analyst]
---

Break revenue down by country.`,
			"skills/conventions/SKILL.md": `---
description: Modeling conventions.
always_apply: true
---

Always materialize models.`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)
	s := newSession(t, rt, instanceID)

	skills, err := s.Skills(t.Context())
	require.NoError(t, err)
	require.NoError(t, ai.PreloadSkills(t.Context(), s, skills, parser.SkillAgentAnalyst))

	require.Len(t, s.Messages(ai.FilterByType(ai.MessageTypeCall), ai.FilterByTool(ai.ListSkillsName)), 1)
	loads := s.Messages(ai.FilterByType(ai.MessageTypeCall), ai.FilterByTool(ai.LoadSkillName))
	require.Len(t, loads, 1)
	require.Contains(t, loads[0].Content, `"glossary"`)
	// The developer-only always-apply skill and the on-demand skill are not loaded
	for _, m := range s.Messages(ai.FilterByType(ai.MessageTypeResult), ai.FilterByTool(ai.LoadSkillName)) {
		require.NotContains(t, m.Content, "Always materialize models.")
		require.NotContains(t, m.Content, "Break revenue down by country.")
	}
}

// TestSkillsEmptyProject verifies that the skill tools stay available in a project without skills,
// so that MCP clients see the same tools for every project, and that they return recoverable results.
func TestSkillsEmptyProject(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	s := newSession(t, rt, instanceID)

	for _, name := range []string{ai.ListSkillsName, ai.LoadSkillName} {
		tool, ok := s.Tool(name)
		require.True(t, ok)
		allowed, err := tool.CheckAccess(ai.WithSession(t.Context(), s))
		require.NoError(t, err)
		require.True(t, allowed, "tool %q access", name)
	}

	var res *ai.ListSkillsResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListSkillsName, &res, &ai.ListSkillsArgs{})
	require.NoError(t, err)
	require.Empty(t, res.Skills)

	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.LoadSkillName, nil, &ai.LoadSkillArgs{Name: "anything"})
	require.ErrorContains(t, err, "does not define any skills")
}

// TestSkillsValidation verifies that invalid skill files surface as parse errors on the file,
// and don't affect the valid skills in the project.
func TestSkillsValidation(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"skills/valid/SKILL.md": `---
description: A valid skill.
---

Body.`,
			"skills/broken/SKILL.md": `# No front matter here`,
		},
	})

	ctrl, err := rt.Controller(t.Context(), instanceID)
	require.NoError(t, err)
	pp, err := ctrl.Get(t.Context(), runtime.GlobalProjectParserName, false)
	require.NoError(t, err)
	parseErrors := pp.GetProjectParser().State.ParseErrors
	require.Len(t, parseErrors, 1)
	require.Equal(t, "/skills/broken/SKILL.md", parseErrors[0].FilePath)
	require.Contains(t, parseErrors[0].Message, "front matter")

	s := newSession(t, rt, instanceID)
	var listRes *ai.ListSkillsResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.ListSkillsName, &listRes, &ai.ListSkillsArgs{})
	require.NoError(t, err)
	require.Len(t, listRes.Skills, 1)
	require.Equal(t, "valid", listRes.Skills[0].Name)
}

// TestSkillsInListMetricsViews verifies that always-apply skills are appended to the
// ai_instructions field of list_metrics_views for external MCP clients,
// but not for Rill's own agents (which receive them directly in their prompts).
func TestSkillsInListMetricsViews(t *testing.T) {
	// The project instructions alone fill the always-apply budget, which must not crowd out the skills.
	longInstructions := "Revenue always refers to net revenue. " + strings.Repeat("x", 1<<15)
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml":         "ai_instructions: " + longInstructions,
			"models/orders.sql": `SELECT 1 AS revenue`,
			"metrics/orders.yaml": `
type: metrics_view
model: orders
measures:
- name: revenue
  expression: SUM(revenue)
`,
			"skills/glossary/SKILL.md": `---
description: Business glossary.
agents: [analyst]
always_apply: true
---

ARPU excludes trial users.`,
			"skills/on-demand/SKILL.md": `---
description: An on-demand skill.
agents: [analyst]
---

Not injected wholesale.`,
			// Scoped always-apply skill: injected with its scope, since the tool has no selected metrics view
			"skills/orders-rca/SKILL.md": `---
description: Orders playbook.
metrics_views: [orders]
agents: [analyst]
always_apply: true
---

Break revenue down by country.`,
			// Developer skill (the default agent): not relevant to external analysis clients
			"skills/dev-conventions/SKILL.md": `---
description: Modeling conventions.
always_apply: true
---

Always materialize models.`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 8, 0, 0)

	newSession := func(t *testing.T, userAgent string, claims *runtime.SecurityClaims) *ai.Session {
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

	// External MCP client: always-apply skill bodies are included after the project instructions, on-demand skill is not
	s := newSession(t, "mcp-client", &runtime.SecurityClaims{UserID: uuid.NewString(), SkipChecks: true})
	var res *ai.ListMetricsViewsResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListMetricsViewsName, &res, &ai.ListMetricsViewsArgs{})
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(res.AIInstructions, longInstructions))
	require.Contains(t, res.AIInstructions, "## Skill: glossary\n\nARPU excludes trial users.")
	require.Contains(t, res.AIInstructions, "## Skill: orders-rca\n\nApplies to the metrics views: orders.\n\nBreak revenue down by country.")
	require.NotContains(t, res.AIInstructions, "Not injected wholesale.")
	require.NotContains(t, res.AIInstructions, "## Skill: dev-conventions")

	// Rill's own agents: no enrichment (skills are injected into their prompts instead)
	s = newSession(t, "rill-web", &runtime.SecurityClaims{UserID: uuid.NewString(), SkipChecks: true})
	res = nil
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.ListMetricsViewsName, &res, &ai.ListMetricsViewsArgs{})
	require.NoError(t, err)
	require.Empty(t, res.AIInstructions)
}

// TestSkillsMCPAccess verifies that the skill tools are exposed to any principal with UseAI,
// including viewers without repo access.
func TestSkillsMCPAccess(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"skills/glossary/SKILL.md": `---
description: Business glossary.
---

ARPU excludes trial users.`,
		},
	})

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
