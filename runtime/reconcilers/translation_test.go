package reconcilers_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

const translationTestModel = `SELECT 'foo' as publisher, 1 as x`

const translationTestMetricsView = `
version: 1
type: metrics_view
model: m1
display_name: Auctions
dimensions:
- name: publisher
  column: publisher
measures:
- name: total_bids
  expression: sum(x)
`

const translationTestExplore = `
type: explore
display_name: Auction explorer
metrics_view: mv1
dimensions: '*'
measures: '*'
`

func TestTranslations(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"models/m1.sql":          translationTestModel,
		"metrics_views/mv1.yaml": translationTestMetricsView,
		"explores/e1.yaml":       translationTestExplore,
		"locales/fr.json": `{
			"mv1": {
				"display_name": "Enchères",
				"dimensions": { "publisher": { "display_name": "Éditeur" } },
				"measures": { "total_bids": { "display_name": "Total des enchères" } }
			},
			"e1": { "display_name": "Explorateur d'enchères" }
		}`,
	})

	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, -1, 0, 0)
	testruntime.RequireResource(t, rt, id, &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name: &runtimev1.ResourceName{Kind: runtime.ResourceKindTranslation, Name: "fr"},
			Refs: []*runtimev1.ResourceName{
				{Kind: runtime.ResourceKindMetricsView, Name: "mv1"},
				{Kind: runtime.ResourceKindExplore, Name: "e1"},
			},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/locales/fr.json"},
		},
		Resource: &runtimev1.Resource_Translation{
			Translation: &runtimev1.Translation{
				Spec: &runtimev1.TranslationSpec{
					Resources: map[string]*runtimev1.TranslationSpec_ResourceTranslation{
						"mv1": {
							BaseTranslation: &runtimev1.TranslationSpec_Labels{
								Type:   runtimev1.TranslationSpec_LABELS_TYPE_BASE,
								Labels: map[string]string{"display_name": "Enchères"},
							},
							SubTranslations: map[string]*runtimev1.TranslationSpec_Labels{
								"publisher": {
									Type:   runtimev1.TranslationSpec_LABELS_TYPE_DIMENSION,
									Labels: map[string]string{"display_name": "Éditeur"},
								},
								"total_bids": {
									Type:   runtimev1.TranslationSpec_LABELS_TYPE_MEASURE,
									Labels: map[string]string{"display_name": "Total des enchères"},
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
				},
				State: &runtimev1.TranslationState{},
			},
		},
	})

	// An unknown locale doesn't resolve to any translations, and applying nil is a no-op.
	translations, err := rt.TranslationsForLocale(t.Context(), id, "de")
	require.NoError(t, err)
	require.Nil(t, translations)

	translations, err = rt.TranslationsForLocale(t.Context(), id, "fr")
	require.NoError(t, err)
	require.NotNil(t, translations)

	// Both the spec and the valid spec are translated.
	mv := runtime.ApplyTranslations(testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1"), translations).GetMetricsView()
	require.Equal(t, "Enchères", mv.Spec.DisplayName)
	require.Equal(t, "Enchères", mv.State.ValidSpec.DisplayName)
	require.Equal(t, "Éditeur", mv.State.ValidSpec.Dimensions[0].DisplayName)
	require.Equal(t, "Total des enchères", mv.State.ValidSpec.Measures[0].DisplayName)

	exp := runtime.ApplyTranslations(testruntime.GetResource(t, rt, id, runtime.ResourceKindExplore, "e1"), translations).GetExplore()
	require.Equal(t, "Explorateur d'enchères", exp.Spec.DisplayName)
	require.Equal(t, "Explorateur d'enchères", exp.State.ValidSpec.DisplayName)

	// The untranslated resource is left alone.
	require.Equal(t, "Auctions", testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1").GetMetricsView().Spec.DisplayName)
}

func TestTranslationsValidation(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr string
	}{
		{
			name:    "unknown resource",
			json:    `{"nope": {"display_name": "Enchères"}}`,
			wantErr: `"nope" does not match a translatable resource`,
		},
		{
			name:    "name differs in case",
			json:    `{"MV1": {"display_name": "Enchères"}}`,
			wantErr: `"MV1" does not match the name of MetricsView/mv1 exactly`,
		},
		{
			name:    "unknown dimension",
			json:    `{"mv1": {"dimensions": {"nope": {"display_name": "Éditeur"}}}}`,
			wantErr: `"mv1": "nope" is not a dimension of the metrics view`,
		},
		{
			name:    "measure filed under dimensions",
			json:    `{"mv1": {"dimensions": {"total_bids": {"display_name": "Total"}}}}`,
			wantErr: `"mv1": "total_bids" is not a dimension of the metrics view`,
		},
		{
			name:    "sub translations on an explore",
			json:    `{"e1": {"dimensions": {"publisher": {"display_name": "Éditeur"}}}}`,
			wantErr: `"e1": dimensions and measures can't be translated on Explore`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rt, id := testruntime.NewInstance(t)
			testruntime.PutFiles(t, rt, id, map[string]string{
				"models/m1.sql":          translationTestModel,
				"metrics_views/mv1.yaml": translationTestMetricsView,
				"explores/e1.yaml":       translationTestExplore,
				"locales/fr.json":        test.json,
			})

			testruntime.ReconcileParserAndWait(t, rt, id)
			testruntime.RequireReconcileState(t, rt, id, -1, 1, 0)

			res := testruntime.GetResource(t, rt, id, runtime.ResourceKindTranslation, "fr")
			require.Contains(t, res.Meta.ReconcileError, test.wantErr)
		})
	}
}
