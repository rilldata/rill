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
				Variables:                []string{},
				UsesTemplating:           false,
				ResolvedWithPlaceholders: `SELECT * FROM foo`,
			},
		},
		{
			name:     "ref",
			template: `SELECT * FROM {{ ref "foo" }}`,
			want: &TemplateMetadata{
				Refs:                     []ResourceName{{Name: "foo"}},
				Variables:                []string{},
				UsesTemplating:           true,
				ResolvedWithPlaceholders: `SELECT * FROM <no value>`,
			},
		},
		{
			name:     "configure",
			template: `{{ configure "a" "b" }}SELECT * FROM foo`,
			want: &TemplateMetadata{
				Config:                   map[string]any{"a": "b"},
				Variables:                []string{},
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
				Variables:                []string{"env.world"},
				UsesTemplating:           true,
				ResolvedWithPlaceholders: `SELECT * FROM <no value> WHERE hello='<no value>' AND world='<no value>'`,
			},
		},
		{
			name:     "variables",
			template: `SELECT * FROM {{.env.partner_table_name}} WITH SAMPLING {{.env.partner_table_name}} .... {{.user.domain}}`,
			want: &TemplateMetadata{
				Refs:                     []ResourceName{},
				Config:                   map[string]any{},
				Variables:                []string{"env.partner_table_name", "user.domain"},
				UsesTemplating:           true,
				ResolvedWithPlaceholders: `SELECT * FROM <no value> WITH SAMPLING <no value> .... <no value>`,
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
				require.ElementsMatch(t, tc.want.Variables, got.Variables)
				require.Equal(t, tc.want.Config, got.Config)
				require.Equal(t, tc.want.UsesTemplating, got.UsesTemplating)
				require.Equal(t, tc.want.ResolvedWithPlaceholders, strings.TrimSpace(got.ResolvedWithPlaceholders))
			}
		})
	}
}

func TestResolve(t *testing.T) {
	template := "SELECT partner_id FROM domain_partner_mapping WHERE domain = '{{ .user.domain }}' AND groups IN ('{{ .user.groups | join \"', '\" }}') {{ if development }}OR true{{ end }}"
	resolved, err := ResolveTemplate(template, TemplateData{
		Environment: "development",
		User: map[string]any{
			"domain": "rilldata.com",
			"groups": []string{"admin", "user"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "SELECT partner_id FROM domain_partner_mapping WHERE domain = 'rilldata.com' AND groups IN ('admin', 'user') OR true", resolved)
}
