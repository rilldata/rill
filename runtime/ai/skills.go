package ai

import (
	"context"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/instructions"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

// skillsGlob matches skill files in the project repo.
// Skills are markdown files with YAML front matter, located at `skills/<name>.md` or `skills/<name>/SKILL.md`.
const skillsGlob = "skills/**/*.md"

// skillMaxFileSize is the maximum size of a skill file.
// It matches the parser's per-file limit so skill files remain valid if they become parsed resources in the future.
const skillMaxFileSize = 1 << 17 // 128kb

// skillsMaxAlwaysApplyBytes caps the total size of always-apply skill bodies injected into a prompt.
// Skills that exceed the cap fall back to on-demand loading via the load_skill tool.
const skillsMaxAlwaysApplyBytes = 1 << 15 // 32kb

// Agents that a skill can target via the `agents` front matter field.
const (
	skillAgentAnalyst   = "analyst"
	skillAgentDeveloper = "developer"
)

// Skill is a user-defined instruction file that teaches Rill's AI agents project-specific practices,
// such as analysis playbooks, business glossaries, or development conventions.
type Skill struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Description  string   `json:"description"`
	MetricsViews []string `json:"metrics_views,omitempty"`
	Agents       []string `json:"agents"`
	AlwaysApply  bool     `json:"always_apply"`
	Body         string   `json:"body"`
}

// SkillIssue describes a skill file that could not be loaded.
type SkillIssue struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

// skillFrontMatter is the YAML front matter of a skill file.
type skillFrontMatter struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	MetricsViews []string `yaml:"metrics_views"`
	Agents       []string `yaml:"agents"`
	AlwaysApply  bool     `yaml:"always_apply"`
}

// Skills lazily loads the project's skill files, memoizing the result for the lifetime of the session.
// Malformed skill files are reported as issues and logged; they never fail the load as a whole.
func (s *BaseSession) Skills(ctx context.Context) ([]*Skill, []SkillIssue, error) {
	s.skillsMu.Lock()
	defer s.skillsMu.Unlock()
	if s.skillsLoaded {
		return s.skills, s.skillIssues, nil
	}

	skills, issues, err := loadSkills(ctx, s.runner.Runtime, s.instanceID)
	if err != nil {
		return nil, nil, err
	}
	for _, issue := range issues {
		s.logger.Warn("skipping invalid skill file", zap.String("path", issue.Path), zap.String("error", issue.Error))
	}

	s.skills = skills
	s.skillIssues = issues
	s.skillsLoaded = true
	return s.skills, s.skillIssues, nil
}

// loadSkills reads and parses all skill files in the project repo.
func loadSkills(ctx context.Context, rt *runtime.Runtime, instanceID string) ([]*Skill, []SkillIssue, error) {
	repo, release, err := rt.Repo(ctx, instanceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open repo: %w", err)
	}
	defer release()

	entries, err := repo.ListGlob(ctx, skillsGlob, true)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list skill files: %w", err)
	}

	// Sort by path so duplicate name resolution is deterministic (first path wins).
	slices.SortFunc(entries, func(a, b drivers.DirEntry) int { return strings.Compare(a.Path, b.Path) })

	var skills []*Skill
	var issues []SkillIssue
	seen := map[string]string{} // skill name to the path that claimed it
	for _, entry := range entries {
		name, ok := skillNameForPath(entry.Path)
		if !ok {
			// E.g. an auxiliary markdown file inside a skill directory
			continue
		}

		content, err := repo.Get(ctx, entry.Path)
		if err != nil {
			issues = append(issues, SkillIssue{Path: entry.Path, Error: fmt.Sprintf("failed to read file: %s", err)})
			continue
		}
		if len(content) > skillMaxFileSize {
			issues = append(issues, SkillIssue{Path: entry.Path, Error: fmt.Sprintf("file exceeds the maximum skill size of %d bytes", skillMaxFileSize)})
			continue
		}

		var fm skillFrontMatter
		body, err := instructions.ParseFrontMatter([]byte(content), &fm)
		if err != nil {
			issues = append(issues, SkillIssue{Path: entry.Path, Error: err.Error()})
			continue
		}
		if fm.Name != "" {
			name = fm.Name
		}
		if fm.Description == "" {
			issues = append(issues, SkillIssue{Path: entry.Path, Error: "missing required \"description\" property in front matter"})
			continue
		}
		if prev, ok := seen[name]; ok {
			issues = append(issues, SkillIssue{Path: entry.Path, Error: fmt.Sprintf("duplicate skill name %q (already defined in %q)", name, prev)})
			continue
		}
		seen[name] = entry.Path

		agents := fm.Agents
		if len(agents) == 0 {
			agents = []string{skillAgentAnalyst}
		}

		skills = append(skills, &Skill{
			Name:         name,
			Path:         entry.Path,
			Description:  fm.Description,
			MetricsViews: fm.MetricsViews,
			Agents:       agents,
			AlwaysApply:  fm.AlwaysApply,
			Body:         body,
		})
	}

	return skills, issues, nil
}

// skillNameForPath derives a skill's default name from its file path.
// Valid skill paths are `/skills/<name>.md` and `/skills/<name>/SKILL.md`;
// other markdown files under `/skills/` (such as auxiliary files in a skill directory) return false.
func skillNameForPath(p string) (string, bool) {
	parts := strings.Split(strings.TrimPrefix(path.Clean(p), "/"), "/")
	if len(parts) < 2 || parts[0] != "skills" {
		return "", false
	}
	switch len(parts) {
	case 2:
		return strings.TrimSuffix(parts[1], ".md"), true
	case 3:
		if parts[2] == "SKILL.md" {
			return parts[1], true
		}
	}
	return "", false
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
