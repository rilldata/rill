package ai

import (
	"context"
	"slices"
	"strings"

	"github.com/rilldata/rill/runtime"
)

// skillsMaxAlwaysApplyBytes caps the total size of always-apply skill bodies injected into a prompt.
// Skills that exceed the cap fall back to on-demand loading via the load_skill tool.
const skillsMaxAlwaysApplyBytes = 1 << 15 // 32kb

// Skill is a user-defined instruction file that teaches Rill's AI agents project-specific practices,
// such as analysis playbooks, business glossaries, or development conventions.
// Skills are parsed from SKILL.md files into catalog resources; see runtime/parser/parse_skill.go.
type Skill struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Description  string   `json:"description"`
	MetricsViews []string `json:"metrics_views,omitempty"`
	Agents       []string `json:"agents"`
	AlwaysApply  bool     `json:"always_apply"`
	Body         string   `json:"body"`
}

// Skills lazily loads the project's skills from the catalog, memoizing the result for the lifetime of the session.
func (s *BaseSession) Skills(ctx context.Context) ([]*Skill, error) {
	s.skillsMu.Lock()
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
		spec := r.GetSkill().Spec
		var path string
		if len(r.Meta.FilePaths) > 0 {
			path = r.Meta.FilePaths[0]
		}
		skills = append(skills, &Skill{
			Name:         r.Meta.Name.Name,
			Path:         path,
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

// filterSkills returns the skills relevant to the given agent and metrics view context.
// A skill scoped to specific metrics views is included only if the context references one of them.
// An empty context includes all of the agent's skills: scoping is a relevance filter, not access control.
func filterSkills(skills []*Skill, agent string, metricsViewNames []string) []*Skill {
	var res []*Skill
	for _, sk := range skills {
		if !slices.Contains(sk.Agents, agent) {
			continue
		}
		if len(sk.MetricsViews) > 0 && len(metricsViewNames) > 0 {
			relevant := slices.ContainsFunc(sk.MetricsViews, func(mv string) bool {
				return slices.Contains(metricsViewNames, mv)
			})
			if !relevant {
				continue
			}
		}
		res = append(res, sk)
	}
	return res
}
