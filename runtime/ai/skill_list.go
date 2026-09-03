package ai

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const ListSkillsName = "list_skills"

type ListSkills struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ListSkillsArgs, *ListSkillsResult] = (*ListSkills)(nil)

type ListSkillsArgs struct{}

type ListSkillsResult struct {
	Skills []*SkillInfo `json:"skills"`
}

// SkillInfo describes a skill without its body. Use load_skill to fetch the full instructions.
type SkillInfo struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	MetricsViews []string `json:"metrics_views,omitempty"`
	Agents       []string `json:"agents"`
	AlwaysApply  bool     `json:"always_apply"`
}

func (t *ListSkills) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ListSkillsName,
		Title:       "List skills",
		Description: "Lists the skills defined in the project. Skills are instruction files that teach AI agents project-specific analysis or development practices. Load a skill with the load_skill tool before doing work its description covers. Treat skills marked always_apply as standing instructions and load them up front.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    true,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Listing skills...",
			"openai/toolInvocation/invoked":  "Listed skills",
		},
	}
}

func (t *ListSkills) CheckAccess(ctx context.Context) (bool, error) {
	s := GetSession(ctx)
	return s.Claims().Can(runtime.UseAI), nil
}

func (t *ListSkills) Handler(ctx context.Context, args *ListSkillsArgs) (*ListSkillsResult, error) {
	s := GetSession(ctx)

	skills, err := s.Skills(ctx)
	if err != nil {
		return nil, err
	}

	infos := make([]*SkillInfo, len(skills))
	for i, sk := range skills {
		infos[i] = &SkillInfo{
			Name:         sk.Name,
			Description:  sk.Description,
			MetricsViews: sk.MetricsViews,
			Agents:       sk.Agents,
			AlwaysApply:  sk.AlwaysApply,
		}
	}

	return &ListSkillsResult{Skills: infos}, nil
}
