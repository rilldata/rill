package ai

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/rilldata/rill/runtime"
	"go.uber.org/zap"
)

// skillsMaxAlwaysApplyBytes caps the total size of always-apply skill bodies injected into a prompt.
// Skills that exceed the cap fall back to on-demand loading via the load_skill tool.
const skillsMaxAlwaysApplyBytes = 1 << 15 // 32kb

// skillsMaxIndexBytes caps the size of the skill index injected into a prompt.
// Skills that don't fit remain discoverable via the list_skills tool.
const skillsMaxIndexBytes = 1 << 14 // 16kb

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
// They are only exposed when the project defines skills, so clients of projects without skills never see them.
func checkSkillAccess(ctx context.Context) (bool, error) {
	s := GetSession(ctx)
	if !s.Claims().Can(runtime.UseAI) {
		return false, nil
	}
	skills, err := s.Skills(ctx)
	if err != nil {
		return false, err
	}
	return len(skills) > 0, nil
}

// filterSkills returns the skills relevant to the given agent and metrics view context.
// A skill scoped to specific metrics views is included only if the context references one of them.
// Metrics view names are compared case-insensitively, matching how the catalog identifies resources.
// An empty context includes all of the agent's skills: scoping is a relevance filter, not access control.
func filterSkills(skills []*Skill, agent string, metricsViewNames []string) []*Skill {
	var res []*Skill
	for _, sk := range skills {
		if !slices.Contains(sk.Agents, agent) {
			continue
		}
		if len(sk.MetricsViews) > 0 && len(metricsViewNames) > 0 {
			relevant := slices.ContainsFunc(sk.MetricsViews, func(mv string) bool {
				return slices.ContainsFunc(metricsViewNames, func(name string) bool { return strings.EqualFold(name, mv) })
			})
			if !relevant {
				continue
			}
		}
		res = append(res, sk)
	}
	return res
}

// skillSection renders an always-apply skill as a section for inclusion in a prompt or in ai_instructions.
// A skill scoped to metrics views states its scope, since it may be injected where no metrics view has been selected yet.
func skillSection(sk *Skill) string {
	var scope string
	if len(sk.MetricsViews) > 0 {
		scope = fmt.Sprintf("Applies to the metrics views: %s.\n\n", strings.Join(sk.MetricsViews, ", "))
	}
	return fmt.Sprintf("## Skill: %s\n\n%s%s", sk.Name, scope, sk.Body)
}

// skillPrompts splits skills into the always-apply bodies to inject into an agent's prompt wholesale
// and an index of the remaining skills for the agent to fetch on demand with the load_skill tool.
// An always-apply body that would exceed the size cap falls back to the on-demand index.
// The index is capped too; skills that don't fit are counted and the agent is pointed to the list_skills tool.
func skillPrompts(skills []*Skill, logger *zap.Logger) (alwaysApply, index string) {
	var alwaysApplyBuf, indexBuf strings.Builder
	var omitted int
	for _, sk := range skills {
		if sk.AlwaysApply {
			// The cap applies to the rendered section, including its heading, not just the body.
			section := skillSection(sk) + "\n\n"
			if alwaysApplyBuf.Len()+len(section) <= skillsMaxAlwaysApplyBytes {
				alwaysApplyBuf.WriteString(section)
				continue
			}
			logger.Warn("always-apply skill exceeds the prompt size cap; falling back to on-demand loading", zap.String("skill", sk.Name))
		}
		entry := fmt.Sprintf("- %s: %s\n", sk.Name, sk.Description)
		if indexBuf.Len()+len(entry) > skillsMaxIndexBytes {
			omitted++
			continue
		}
		indexBuf.WriteString(entry)
	}
	if omitted > 0 {
		logger.Warn("skill index exceeds the prompt size cap; some skills are only discoverable with list_skills", zap.Int("omitted", omitted))
		fmt.Fprintf(&indexBuf, "- (%d more skills not listed here; call %s to see them)\n", omitted, ListSkillsName)
	}
	return strings.TrimSpace(alwaysApplyBuf.String()), strings.TrimSpace(indexBuf.String())
}
