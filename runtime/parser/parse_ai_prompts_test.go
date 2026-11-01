package parser

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestParseAIPrompts(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    []*runtimev1.AIPrompt
		wantErr string
	}{
		{
			name: "string and mapping forms",
			yaml: `
ai_prompts:
  - Which campaigns drove the biggest change in impressions this week?
  - label: CTR outliers
    prompt: Which publishers have a click-through rate far above or below the average?
  - prompt: "  Short.  "
`,
			want: []*runtimev1.AIPrompt{
				{Label: "Which campaigns drove the biggest change…", Prompt: "Which campaigns drove the biggest change in impressions this week?"},
				{Label: "CTR outliers", Prompt: "Which publishers have a click-through rate far above or below the average?"},
				{Label: "Short", Prompt: "Short."},
			},
		},
		{
			name:    "empty prompt",
			yaml:    "ai_prompts:\n  - label: Foo\n",
			wantErr: "ai_prompts entry 1 must have a non-empty prompt",
		},
		{
			name:    "duplicate prompt",
			yaml:    "ai_prompts:\n  - What is the total?\n  - label: Again\n    prompt: 'What is the total? '\n",
			wantErr: "ai_prompts entry 2 duplicates an earlier prompt",
		},
		{
			name:    "too many prompts",
			yaml:    "ai_prompts: [a, b, c, d, e, f, g, h, i]\n",
			wantErr: "ai_prompts can have at most 8 entries",
		},
		{
			name:    "label too long",
			yaml:    "ai_prompts:\n  - label: " + string(repeatByte('x', 41)) + "\n    prompt: foo\n",
			wantErr: "maximum label length",
		},
		{
			name:    "invalid shape",
			yaml:    "ai_prompts:\n  - [foo]\n",
			wantErr: "an AI prompt must be a string or a mapping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := makeRepo(t, map[string]string{
				"rill.yaml":        "",
				"explores/e1.yaml": "type: explore\nmetrics_view: mv1\n" + tt.yaml,
				"canvases/c1.yaml": "type: canvas\n" + tt.yaml,
			})
			p, err := Parse(context.Background(), repo, "", "", "duckdb", true)
			require.NoError(t, err)
			if tt.wantErr != "" {
				require.Len(t, p.Errors, 2)
				for _, e := range p.Errors {
					require.Contains(t, e.Message, tt.wantErr)
				}
				return
			}
			require.Empty(t, p.Errors)
			var explore *runtimev1.ExploreSpec
			var canvas *runtimev1.CanvasSpec
			for _, resource := range p.Resources {
				switch resource.Name.Kind {
				case ResourceKindExplore:
					explore = resource.ExploreSpec
				case ResourceKindCanvas:
					canvas = resource.CanvasSpec
				}
			}
			require.NotNil(t, explore)
			require.NotNil(t, canvas)
			requireAIPromptsEqual(t, tt.want, explore.AiPrompts)
			requireAIPromptsEqual(t, tt.want, canvas.AiPrompts)
		})
	}
}

func TestParseAIPromptsInRillYAML(t *testing.T) {
	repo := makeRepo(t, map[string]string{
		"rill.yaml": `
ai_prompts:
  - What data is available in this project?
  - label: Key metrics
    prompt: Give me an overview of the key metrics across the project.
`,
	})
	p, err := Parse(context.Background(), repo, "", "", "duckdb", true)
	require.NoError(t, err)
	require.Empty(t, p.Errors)
	requireAIPromptsEqual(t, []*runtimev1.AIPrompt{
		{Label: "What data is available in this…", Prompt: "What data is available in this project?"},
		{Label: "Key metrics", Prompt: "Give me an overview of the key metrics across the project."},
	}, p.RillYAML.AIPrompts)

	repo = makeRepo(t, map[string]string{
		"rill.yaml": "ai_prompts:\n  - prompt: ''\n",
	})
	p, err = Parse(context.Background(), repo, "", "", "duckdb", true)
	require.NoError(t, err)
	require.Len(t, p.Errors, 1)
	require.Contains(t, p.Errors[0].Message, "non-empty prompt")
}

func requireAIPromptsEqual(t *testing.T, want, got []*runtimev1.AIPrompt) {
	t.Helper()
	require.Len(t, got, len(want))
	for i := range want {
		require.Equal(t, want[i].Label, got[i].Label, "label of prompt %d", i)
		require.Equal(t, want[i].Prompt, got[i].Prompt, "prompt of prompt %d", i)
	}
}

func repeatByte(b byte, n int) []byte {
	res := make([]byte, n)
	for i := range res {
		res[i] = b
	}
	return res
}
