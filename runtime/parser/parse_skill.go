package parser

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gopkg.in/yaml.v3"
)

// Skills are directories containing a SKILL.md file that follows the Agent Skills format (https://agentskills.io).
// They teach AI agents project-specific practices, such as analysis playbooks and business glossaries.
// Rill loads skills from the `skills` directory, and additionally from the generic `.agents/skills` directory
// for compatibility with skills authored for other agent clients.

// Rill agents that a skill can target via the `agents` front matter field.
const (
	SkillAgentAnalyst   = "analyst"
	SkillAgentDeveloper = "developer"
)

// skillYAML is the front matter of a SKILL.md file.
// The `license`, `compatibility`, `metadata` and `allowed-tools` fields are defined by the Agent Skills format;
// Rill accepts them for compatibility but does not currently use them.
type skillYAML struct {
	// Fields defined by the Agent Skills format
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	License       string            `yaml:"license"`
	Compatibility string            `yaml:"compatibility"`
	Metadata      map[string]string `yaml:"metadata"`
	AllowedTools  string            `yaml:"allowed-tools"`
	// Rill extensions
	MetricsViews []string `yaml:"metrics_views"`
	Agents       []string `yaml:"agents"`
	AlwaysApply  bool     `yaml:"always_apply"`
}

// skillNameRegexp validates skill names per the Agent Skills format:
// lowercase alphanumeric characters and single hyphens, not at the start or end.
var skillNameRegexp = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// parseSkill parses a SKILL.md file and adds the resulting resource to p.Resources.
func (p *Parser) parseSkill(ctx context.Context, path string) error {
	name, ok := skillNameForPath(path)
	if !ok {
		panic(fmt.Errorf("parseSkill called with invalid skill path %q", path))
	}

	data, err := p.Repo.Get(ctx, path)
	if err != nil {
		if os.IsNotExist(err) {
			// This is a dirty parse where a file disappeared during parsing.
			// Due to the clear-and-rebuild behavior, we can safely continue parsing.
			return nil
		}
		return err
	}
	if len(data) > maxFileSize {
		return fmt.Errorf("size %d bytes exceeds max size of %d bytes", len(data), maxFileSize)
	}

	tmp := &skillYAML{}
	body, err := parseSkillFrontMatter(data, tmp)
	if err != nil {
		return err
	}

	// Per the Agent Skills format, a `name` in the front matter must match the skill's directory name.
	// Rill is lenient and derives the name from the directory when the front matter omits it.
	if tmp.Name != "" && tmp.Name != name {
		return fmt.Errorf("front matter name %q must match the skill's directory name %q", tmp.Name, name)
	}
	if len(name) > 64 || !skillNameRegexp.MatchString(name) {
		return fmt.Errorf("invalid skill name %q: names must be at most 64 characters of lowercase letters, numbers and non-consecutive hyphens", name)
	}
	if tmp.Description == "" {
		return errors.New(`missing required front matter field "description"`)
	}
	if len(tmp.Description) > 1024 {
		return fmt.Errorf(`front matter field "description" exceeds the maximum length of 1024 characters`)
	}

	agents := tmp.Agents
	if len(agents) == 0 {
		agents = []string{SkillAgentAnalyst}
	}
	for _, agent := range agents {
		if agent != SkillAgentAnalyst && agent != SkillAgentDeveloper {
			return fmt.Errorf("invalid agent %q: expected %q or %q", agent, SkillAgentAnalyst, SkillAgentDeveloper)
		}
	}

	// Reference the metrics views the skill is scoped to, so a typo surfaces as an error on the file
	// and the skill is revisited when the metrics views change.
	refs := make([]ResourceName, len(tmp.MetricsViews))
	for i, mv := range tmp.MetricsViews {
		refs[i] = ResourceName{Kind: ResourceKindMetricsView, Name: mv}
	}

	r, err := p.insertResource(ResourceKindSkill, name, []string{path}, nil, refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, we can't return an error. Must call p.addParseError instead.

	r.SkillSpec = &runtimev1.SkillSpec{
		Description:  tmp.Description,
		Body:         body,
		MetricsViews: tmp.MetricsViews,
		Agents:       agents,
		AlwaysApply:  tmp.AlwaysApply,
	}

	return nil
}

// parseSkillFrontMatter splits a SKILL.md file into YAML front matter and a markdown body,
// strictly decoding the front matter into the provided struct.
func parseSkillFrontMatter(content string, into *skillYAML) (string, error) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return "", errors.New(`skill files must start with YAML front matter delimited by "---" lines`)
	}

	rest := strings.TrimPrefix(strings.TrimPrefix(content, "---"), "\r")[1:] // Skip "---\n" or "---\r\n"
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return "", errors.New(`unclosed front matter: missing closing "---" line`)
	}
	frontMatter := rest[:endIdx]
	body := strings.TrimSpace(rest[endIdx+len("\n---"):])

	if strings.TrimSpace(frontMatter) == "" {
		// Empty front matter; leave the struct zero-valued so required-field validation reports the actual problem
		return body, nil
	}

	dec := yaml.NewDecoder(strings.NewReader(frontMatter))
	dec.KnownFields(true)
	if err := dec.Decode(into); err != nil {
		return "", fmt.Errorf("failed to parse front matter: %w", err)
	}

	return body, nil
}

// pathIsSkill returns true if the path declares a skill.
func pathIsSkill(path string) bool {
	_, ok := skillNameForPath(path)
	return ok
}

// skillNameForPath returns the skill name for a path, or false if the path does not declare a skill.
// Skills are declared by SKILL.md files at `/skills/<name>/SKILL.md` or `/.agents/skills/<name>/SKILL.md`.
func skillNameForPath(path string) (string, bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) == 3 && parts[0] == "skills" && parts[2] == "SKILL.md" {
		return parts[1], true
	}
	if len(parts) == 4 && parts[0] == ".agents" && parts[1] == "skills" && parts[3] == "SKILL.md" {
		return parts[2], true
	}
	return "", false
}
