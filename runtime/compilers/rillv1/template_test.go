package rillv1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnalyze(t *testing.T) {
	tt := []struct {
		name     string
		template string
		want     *TemplateMetadata
		wantErr  error
	}{
		{
			name:     "no template",
			template: `SELECT * FROM foo`,
			want: &TemplateMetadata{
				UsesTemplating:           false,
				ResolvedWithPlaceholders: `SELECT * FROM foo`,
			},
		},
		{
			name:     "ref",
			template: `SELECT * FROM {{ ref "foo" }}`,
			want: &TemplateMetadata{
				Refs:                     []ResourceName{{Name: "foo"}},
				UsesTemplating:           true,
				ResolvedWithPlaceholders: `SELECT * FROM <no value>`,
			},
		},
		{
			name:     "configure",
			template: `{{ configure "a" "b" }}SELECT * FROM foo`,
			want: &TemplateMetadata{
				Config:                   map[string]any{"a": "b"},
				UsesTemplating:           true,
				ResolvedWithPlaceholders: `SELECT * FROM foo`,
			},
		},
		{
			name:     "complex",
			template: `{{ configure "a: b\nc: d" }}{{ configure "e" "f" }}{{ dependency "bar" }} SELECT * FROM {{ ref "model" "foo" }} WHERE hello='{{ .env.world }}' AND world='{{ (lookup "baz").spec.baz.spaz }}'`,
			want: &TemplateMetadata{
				Refs:                     []ResourceName{{Name: "bar"}, {Kind: ResourceKindModel, Name: "foo"}, {Name: "baz"}},
				Config:                   map[string]any{"a": "b", "c": "d", "e": "f"},
				UsesTemplating:           true,
				ResolvedWithPlaceholders: `SELECT * FROM <no value> WHERE hello='<no value>' AND world='<no value>'`,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AnalyzeTemplate(tc.template)
			if tc.wantErr != nil {
				require.EqualError(t, err, tc.wantErr.Error())
			} else {
				require.NoError(t, err)
				if tc.want.Config == nil {
					tc.want.Config = map[string]any{}
				}
				require.ElementsMatch(t, tc.want.Refs, got.Refs)
				require.Equal(t, tc.want.Config, got.Config)
				require.Equal(t, tc.want.UsesTemplating, got.UsesTemplating)
				require.Equal(t, tc.want.ResolvedWithPlaceholders, strings.TrimSpace(got.ResolvedWithPlaceholders))
			}
		})
	}
}
