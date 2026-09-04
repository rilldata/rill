package parser

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

const testTranslationMetricsView = `
version: 1
type: metrics_view
table: t1
dimensions:
  - name: publisher
    column: publisher
measures:
  - name: total_bids
    expression: count(*)
`

const testTranslationExplore = `
type: explore
metrics_view: mv1
`

func TestTranslation(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`:              ``,
		`metrics_views/mv1.yaml`: testTranslationMetricsView,
		`explores/e1.yaml`:       testTranslationExplore,
		`locales/fr.json`: `{
			"mv1": {
				"display_name": "Enchères",
				"description": "Données d'enchères",
				"dimensions": {
					"publisher": { "display_name": "Éditeur" }
				},
				"measures": {
					"total_bids": { "display_name": "Total des enchères", "description": "Nombre d'enchères" }
				}
			},
			"e1": {
				"display_name": "Explorateur d'enchères"
			}
		}`,
	})

	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	require.Empty(t, p.Errors)

	// The resource name is the locale, i.e. the file's stem.
	tr := p.Resources[ResourceName{Kind: ResourceKindTranslation, Name: "fr"}]
	require.NotNil(t, tr)
	require.Equal(t, []string{"/locales/fr.json"}, tr.Paths)

	// The bare names in the file are resolved to the kinds that can be translated.
	require.Equal(t, []ResourceName{
		{Kind: ResourceKindMetricsView, Name: "mv1"},
		{Kind: ResourceKindExplore, Name: "e1"},
	}, tr.Refs)

	// The dimensions and measures blocks merge into one map, distinguished by their type.
	require.Equal(t, &runtimev1.TranslationSpec{
		Resources: map[string]*runtimev1.TranslationSpec_ResourceTranslation{
			"mv1": {
				BaseTranslation: &runtimev1.TranslationSpec_Labels{
					Type: runtimev1.TranslationSpec_LABELS_TYPE_BASE,
					Labels: map[string]string{
						"display_name": "Enchères",
						"description":  "Données d'enchères",
					},
				},
				SubTranslations: map[string]*runtimev1.TranslationSpec_Labels{
					"publisher": {
						Type:   runtimev1.TranslationSpec_LABELS_TYPE_DIMENSION,
						Labels: map[string]string{"display_name": "Éditeur"},
					},
					"total_bids": {
						Type: runtimev1.TranslationSpec_LABELS_TYPE_MEASURE,
						Labels: map[string]string{
							"display_name": "Total des enchères",
							"description":  "Nombre d'enchères",
						},
					},
				},
			},
			"e1": {
				BaseTranslation: &runtimev1.TranslationSpec_Labels{
					Type:   runtimev1.TranslationSpec_LABELS_TYPE_BASE,
					Labels: map[string]string{"display_name": "Explorateur d'enchères"},
				},
			},
		},
	}, tr.TranslationSpec)
}

func TestTranslationErrors(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr string
	}{
		{
			name:    "unknown field",
			json:    `{"mv1": {"displayname": "Enchères"}}`,
			wantErr: `unknown field "displayname"`,
		},
		{
			name:    "misplaced block",
			json:    `{"mv1": {"subs": {"publisher": {"display_name": "Éditeur"}}}}`,
			wantErr: `unknown field "subs"`,
		},
		{
			name:    "malformed json",
			json:    `{"mv1": {"display_name": "Enchères",}}`,
			wantErr: "failed to parse translation file",
		},
		{
			name:    "dimension and measure collision",
			json:    `{"mv1": {"dimensions": {"x": {"display_name": "X"}}, "measures": {"x": {"display_name": "X"}}}}`,
			wantErr: `"mv1": "x" is translated as both a dimension and a measure`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			repo := makeRepo(t, map[string]string{
				`rill.yaml`:              ``,
				`metrics_views/mv1.yaml`: testTranslationMetricsView,
				`locales/fr.json`:        test.json,
			})

			p, err := Parse(ctx, repo, "", "", "duckdb", true)
			require.NoError(t, err)

			require.Len(t, p.Errors, 1)
			require.Equal(t, "/locales/fr.json", p.Errors[0].FilePath)
			require.Contains(t, p.Errors[0].Message, test.wantErr)
			require.NotContains(t, p.Resources, ResourceName{Kind: ResourceKindTranslation, Name: "fr"})
		})
	}
}

func TestTranslationReparse(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`locales/fr.json`: `{
			"mv1": { "display_name": "Enchères" }
		}`,
	})

	// The metrics view doesn't exist yet, so the ref is dropped.
	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	require.Empty(t, p.Errors)
	require.Empty(t, p.Resources[ResourceName{Kind: ResourceKindTranslation, Name: "fr"}].Refs)

	// Adding the metrics view re-infers the ref.
	putRepo(t, repo, map[string]string{`metrics_views/mv1.yaml`: testTranslationMetricsView})
	_, err = p.Reparse(ctx, []string{"/metrics_views/mv1.yaml"})
	require.NoError(t, err)
	require.Empty(t, p.Errors)
	require.Equal(t, []ResourceName{{Kind: ResourceKindMetricsView, Name: "mv1"}}, p.Resources[ResourceName{Kind: ResourceKindTranslation, Name: "fr"}].Refs)

	// Editing the locale file replaces the resource instead of leaving the old one behind.
	putRepo(t, repo, map[string]string{`locales/fr.json`: `{
		"mv1": { "display_name": "Enchères aux annonces" }
	}`})
	diff, err := p.Reparse(ctx, []string{"/locales/fr.json"})
	require.NoError(t, err)
	require.Empty(t, p.Errors)
	require.Equal(t, &Diff{Modified: []ResourceName{{Kind: ResourceKindTranslation, Name: "fr"}}}, diff)
	require.Equal(t,
		"Enchères aux annonces",
		p.Resources[ResourceName{Kind: ResourceKindTranslation, Name: "fr"}].TranslationSpec.Resources["mv1"].BaseTranslation.Labels["display_name"],
	)

	// Deleting the locale file removes the resource.
	require.NoError(t, repo.Delete(ctx, "/locales/fr.json", false))
	diff, err = p.Reparse(ctx, []string{"/locales/fr.json"})
	require.NoError(t, err)
	require.Empty(t, p.Errors)
	require.Equal(t, &Diff{Deleted: []ResourceName{{Kind: ResourceKindTranslation, Name: "fr"}}}, diff)
	require.NotContains(t, p.Resources, ResourceName{Kind: ResourceKindTranslation, Name: "fr"})
}
