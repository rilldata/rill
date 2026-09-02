package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const LoadSkillName = "load_skill"

type LoadSkill struct {
	Runtime *runtime.Runtime
}

var _ Tool[*LoadSkillArgs, *LoadSkillResult] = (*LoadSkill)(nil)

type LoadSkillArgs struct {
	Name string `json:"name" jsonschema:"Name of the skill to load"`
}

type LoadSkillResult struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Body        string `json:"body"`
}

func (t *LoadSkill) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        LoadSkillName,
		Title:       "Using skill",
		Description: "Loads the full instructions for a skill defined in the project. Skills teach project-specific analysis or development practices. Call this before doing work that a skill's description covers, then follow the returned instructions.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    true,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Loading skill...",
			"openai/toolInvocation/invoked":  "Loaded skill",
		},
	}
}

func (t *LoadSkill) CheckAccess(ctx context.Context) (bool, error) {
	s := GetSession(ctx)
	return s.Claims().Can(runtime.UseAI), nil
}

func (t *LoadSkill) Handler(ctx context.Context, args *LoadSkillArgs) (*LoadSkillResult, error) {
	s := GetSession(ctx)

	skills, _, err := s.Skills(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(skills))
	for i, sk := range skills {
		if sk.Name == args.Name {
			return &LoadSkillResult{
				Name:        sk.Name,
				Description: sk.Description,
				Body:        sk.Body,
			}, nil
		}
		names[i] = sk.Name
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("skill %q not found: the project does not define any skills", args.Name)
	}
	return nil, fmt.Errorf("skill %q not found: available skills are %s", args.Name, strings.Join(names, ", "))
}
