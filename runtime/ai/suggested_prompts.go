package ai

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/jsonschema-go/jsonschema"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

// suggestedPromptsTemplateVersion is included in the input hash so that changing the prompt template regenerates prompts for all dashboards.
const suggestedPromptsTemplateVersion = "1"

// suggestedPromptsCount is the number of prompts requested from the LLM and returned to callers.
const suggestedPromptsCount = 4

// suggestedPromptsMaxLabelLength caps the label shown on a prompt's button.
const suggestedPromptsMaxLabelLength = 40

// suggestedPromptsTimeout bounds the LLM call so that a slow provider does not hold up reconciliation for long.
const suggestedPromptsTimeout = 30 * time.Second

// SuggestedPromptsInput describes a dashboard for prompt generation.
// It only carries schema-level metadata (names and descriptions), never data values, so the generated prompts do not leak row-level information.
type SuggestedPromptsInput struct {
	// DashboardKind is "explore" or "canvas".
	DashboardKind string
	DashboardName string
	DisplayName   string
	Description   string
	MetricsViews  []*SuggestedPromptsMetricsView
	// ProjectInstructions are the project-wide `ai_instructions` from rill.yaml.
	ProjectInstructions string
}

// SuggestedPromptsMetricsView is the subset of a metrics view that is relevant for prompt generation.
type SuggestedPromptsMetricsView struct {
	Name           string
	DisplayName    string
	Description    string
	AIInstructions string
	Dimensions     []*SuggestedPromptsField
	Measures       []*SuggestedPromptsField
}

// SuggestedPromptsField is a dimension or measure of a metrics view.
type SuggestedPromptsField struct {
	Name        string
	DisplayName string
	Description string
}

// NewSuggestedPromptsMetricsView builds a SuggestedPromptsMetricsView from a valid metrics view spec.
// If dimensions or measures are non-nil, only the named fields are included (in the given order); otherwise all fields are included.
func NewSuggestedPromptsMetricsView(name string, mv *runtimev1.MetricsViewSpec, dimensions, measures []string) *SuggestedPromptsMetricsView {
	res := &SuggestedPromptsMetricsView{
		Name:           name,
		DisplayName:    mv.DisplayName,
		Description:    mv.Description,
		AIInstructions: mv.AiInstructions,
	}
	for _, d := range mv.Dimensions {
		if dimensions != nil && !slices.Contains(dimensions, d.Name) {
			continue
		}
		res.Dimensions = append(res.Dimensions, &SuggestedPromptsField{Name: d.Name, DisplayName: d.DisplayName, Description: d.Description})
	}
	for _, m := range mv.Measures {
		if measures != nil && !slices.Contains(measures, m.Name) {
			continue
		}
		res.Measures = append(res.Measures, &SuggestedPromptsField{Name: m.Name, DisplayName: m.DisplayName, Description: m.Description})
	}
	return res
}

// SuggestedPromptsHash returns a hash of everything that influences the generated prompts.
// Reconcilers store it alongside the prompts and skip regeneration while it is unchanged.
func SuggestedPromptsHash(in *SuggestedPromptsInput) string {
	h := md5.New()
	write := func(parts ...string) {
		for _, p := range parts {
			_, _ = h.Write([]byte(p))
			_, _ = h.Write([]byte{0})
		}
	}
	write(suggestedPromptsTemplateVersion, in.DashboardKind, in.DashboardName, in.DisplayName, in.Description, in.ProjectInstructions)
	for _, mv := range in.MetricsViews {
		write(mv.Name, mv.DisplayName, mv.Description, mv.AIInstructions)
		for _, d := range mv.Dimensions {
			write("d", d.Name, d.DisplayName, d.Description)
		}
		for _, m := range mv.Measures {
			write("m", m.Name, m.DisplayName, m.Description)
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateSuggestedPrompts asks the instance's AI connector for starter prompts tailored to a dashboard.
// It returns runtime.ErrAINotConfigured (wrapped) if the instance has no AI connector.
func GenerateSuggestedPrompts(ctx context.Context, rt *runtime.Runtime, instanceID string, in *SuggestedPromptsInput) ([]*runtimev1.AIPrompt, error) {
	llm, release, err := rt.AI(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	userPrompt, err := executeTemplate(suggestedPromptsUserTemplate, map[string]any{
		"kind":               in.DashboardKind,
		"name":               in.DashboardName,
		"display_name":       in.DisplayName,
		"description":        in.Description,
		"metrics_views":      in.MetricsViews,
		"ai_instructions":    in.ProjectInstructions,
		"count":              suggestedPromptsCount,
		"max_label_length":   suggestedPromptsMaxLabelLength,
		"has_multiple_views": len(in.MetricsViews) > 1,
		"is_explore":         in.DashboardKind == "explore",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render suggested prompts template: %w", err)
	}

	outputSchema, err := jsonschema.ForType(reflect.TypeFor[suggestedPromptsOutput](), &jsonschema.ForOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to build suggested prompts output schema: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, suggestedPromptsTimeout)
	defer cancel()

	res, err := llm.Complete(ctx, &drivers.CompleteOptions{
		Messages: []*aiv1.CompletionMessage{
			NewTextCompletionMessage(RoleSystem, suggestedPromptsSystemPrompt),
			NewTextCompletionMessage(RoleUser, userPrompt),
		},
		OutputSchema: outputSchema,
	})
	if err != nil {
		return nil, err
	}

	var text strings.Builder
	for _, block := range res.Message.Content {
		if t, ok := block.BlockType.(*aiv1.ContentBlock_Text); ok {
			text.WriteString(t.Text)
		}
	}
	if text.Len() == 0 {
		return nil, errors.New("the AI response did not contain any text")
	}

	var out suggestedPromptsOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(text.String())), &out); err != nil {
		return nil, fmt.Errorf("failed to parse the AI response as suggested prompts: %w", err)
	}

	prompts := make([]*runtimev1.AIPrompt, 0, suggestedPromptsCount)
	for _, p := range out.Prompts {
		prompt := strings.TrimSpace(p.Prompt)
		label := strings.TrimSpace(p.Label)
		if prompt == "" || label == "" {
			continue
		}
		if utf8.RuneCountInString(label) > suggestedPromptsMaxLabelLength {
			label = strings.TrimSpace(string([]rune(label)[:suggestedPromptsMaxLabelLength-1])) + "…"
		}
		prompts = append(prompts, &runtimev1.AIPrompt{Label: label, Prompt: prompt})
		if len(prompts) == suggestedPromptsCount {
			break
		}
	}
	if len(prompts) == 0 {
		return nil, errors.New("the AI response did not contain any usable prompts")
	}
	return prompts, nil
}

// suggestedPromptsOutput is the structured output requested from the LLM.
type suggestedPromptsOutput struct {
	Prompts []suggestedPromptsOutputItem `json:"prompts" jsonschema:"The suggested prompts, in the order they should be displayed."`
}

type suggestedPromptsOutputItem struct {
	Label  string `json:"label" jsonschema:"A short title for the prompt, at most 40 characters, shown on a button."`
	Prompt string `json:"prompt" jsonschema:"The full question the user would ask, written in the first person as one sentence."`
}

const suggestedPromptsSystemPrompt = `You write starter questions for a business intelligence dashboard's AI assistant.
The assistant can query the dashboard's metrics views: it aggregates measures, groups and filters by dimensions, and compares time periods.
Your questions are shown as clickable suggestions to a user who has just opened the dashboard and has not typed anything yet.

Requirements for every question:
- It must be answerable using only the measures and dimensions listed for the dashboard. Never invent fields.
- Refer to fields by their display names in natural language, without code formatting.
- It must be a single, specific, natural sentence written from the user's point of view, and it must not mention time ranges in absolute dates. Use relative phrasing such as "in the selected period" or "compared to the previous period".
- Do not reference filters, segments or values that are not listed.
- Write in the same language as the dashboard's descriptions. Default to English.

Vary the questions across these types, in this order: an overview of the key measures, a breakdown of a measure by an important dimension, a comparison against the previous period, and an anomaly or outlier question.
Prefer measures and dimensions that carry business meaning over technical identifiers.
If instructions from the project administrator are provided, follow their vocabulary and priorities.`

const suggestedPromptsUserTemplate = `Write exactly {{ .count }} starter questions for the following {{ .kind }} dashboard.
Each label must be at most {{ .max_label_length }} characters.

Dashboard: {{ .display_name }}{{ if ne .name .display_name }} (id: {{ .name }}){{ end }}
{{- if .description }}
Description: {{ .description }}
{{- end }}
{{ range .metrics_views }}
## Metrics view: {{ .DisplayName }}{{ if ne .Name .DisplayName }} (id: {{ .Name }}){{ end }}
{{- if .Description }}
{{ .Description }}
{{- end }}
{{- if .AIInstructions }}
Instructions from the administrator for this metrics view:
{{ .AIInstructions }}
{{- end }}

Measures:
{{- range .Measures }}
- {{ .DisplayName }}{{ if .Description }}: {{ .Description }}{{ end }}
{{- end }}

Dimensions:
{{- range .Dimensions }}
- {{ .DisplayName }}{{ if .Description }}: {{ .Description }}{{ end }}
{{- end }}
{{ end }}
{{- if .ai_instructions }}
Instructions from the project administrator, which may or may not be relevant:
{{ .ai_instructions }}
{{- end }}`
