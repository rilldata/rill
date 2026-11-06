package ai

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/rilldata/rill/runtime"
	"go.uber.org/zap"
)

// skillsMaxAlwaysApplyBytes caps the total size of always-apply skill bodies appended to the ai_instructions returned to external MCP clients.
// Skills that exceed the cap must be loaded with the load_skill tool instead.
const skillsMaxAlwaysApplyBytes = 1 << 15 // 32kb

// Skill is a user-defined instruction file that teaches Rill's AI agents project-specific practices,
// such as analysis playbooks, business glossaries, or development conventions.
// Skills are parsed from SKILL.md files into catalog resources; see runtime/parser/parse_skill.go.
type Skill struct {
	Name         string
	Description  string
	MetricsViews []string
	Agents       []string
	AlwaysApply  bool
	Body         string
}

// Skills lazily loads the project's skills from the catalog, memoizing the result for the lifetime of the session.
func (s *BaseSession) Skills(ctx context.Context) ([]*Skill, error) {
	err := s.skillsMu.Lock(ctx)
	if err != nil {
		return nil, err
	}
	defer s.skillsMu.Unlock()
	if s.skillsLoaded {
		return s.skills, nil
	}

	ctrl, err := s.runner.Runtime.Controller(ctx, s.instanceID)
	if err != nil {
		return nil, err
	}

	rs, err := ctrl.List(ctx, runtime.ResourceKindSkill, "", false)
	if err != nil {
		return nil, err
	}

	skills := make([]*Skill, 0, len(rs))
	for _, r := range rs {
		// Skip skills that failed reconciliation (e.g. scoped to a metrics view that doesn't exist),
		// so that the validation actually keeps invalid instructions away from the agents.
		if r.Meta.ReconcileError != "" {
			s.logger.Warn("skipping skill with reconcile error", zap.String("skill", r.Meta.Name.Name), zap.String("error", r.Meta.ReconcileError))
			continue
		}
		spec := r.GetSkill().Spec
		skills = append(skills, &Skill{
			Name:         r.Meta.Name.Name,
			Description:  spec.Description,
			MetricsViews: spec.MetricsViews,
			Agents:       spec.Agents,
			AlwaysApply:  spec.AlwaysApply,
			Body:         spec.Body,
		})
	}
	slices.SortFunc(skills, func(a, b *Skill) int { return strings.Compare(a.Name, b.Name) })

	s.skills = skills
	s.skillsLoaded = true
	return s.skills, nil
}

// checkSkillAccess checks whether the skill tools should be available in the current session.
// They are served even when the project defines no skills, so that external clients see the same tools for every project:
// list_skills then returns an empty list and load_skill a not-found error.
// The in-app agents decide for themselves whether to offer the tools, based on whether the project defines skills.
func checkSkillAccess(ctx context.Context) (bool, error) {
	return GetSession(ctx).Claims().Can(runtime.UseAI), nil
}

// skillsForAgent returns the skills that apply to the given agent.
func skillsForAgent(skills []*Skill, agent string) []*Skill {
	var res []*Skill
	for _, sk := range skills {
		if slices.Contains(sk.Agents, agent) {
			res = append(res, sk)
		}
	}
	return res
}

// preloadSkills seeds the current call with the tool calls that make the project's skills available to an agent:
// a list_skills call so the agent can discover skills and load them on demand, and a load_skill call for each of the agent's always-apply skills.
// The seeded calls become part of the agent's context like any other pre-invoked tool call.
// Tool errors are recorded in the session and don't fail the agent; only context cancellation is returned.
func preloadSkills(ctx context.Context, s *Session, skills []*Skill, agent string) error {
	_, err := s.CallTool(ctx, RoleAssistant, ListSkillsName, nil, &ListSkillsArgs{})
	if err != nil && errors.Is(err, ctx.Err()) {
		return err
	}
	for _, sk := range skillsForAgent(skills, agent) {
		if !sk.AlwaysApply {
			continue
		}
		_, err := s.CallTool(ctx, RoleAssistant, LoadSkillName, nil, &LoadSkillArgs{Name: sk.Name})
		if err != nil && errors.Is(err, ctx.Err()) {
			return err
		}
	}
	return nil
}
