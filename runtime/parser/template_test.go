package parser

import (
	"fmt"
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
	template := "SELECT partner_id FROM domain_partner_mapping WHERE domain = '{{ .user.domain }}' AND groups IN ('{{ .user.groups | join \"', '\" }}') {{ if dev }}OR true{{ end }}"
	resolved, err := ResolveTemplate(template, TemplateData{
		Environment: "dev",
		User: map[string]any{
			"domain": "rilldata.com",
			"groups": []string{"admin", "user"},
		},
	}, false)
	require.NoError(t, err)
	require.Equal(t, "SELECT partner_id FROM domain_partner_mapping WHERE domain = 'rilldata.com' AND groups IN ('admin', 'user') OR true", resolved)
}

func TestVariables(t *testing.T) {
	template := `a={{ .env.a }} b.a={{ .env.b.a }} b.a={{ get .env "b.a" }}`
	resolved, err := ResolveTemplate(template, TemplateData{
		Variables: map[string]string{
			"a":   "1",
			"b.a": "2",
		},
	}, false)
	require.NoError(t, err)
	require.Equal(t, "a=1 b.a=2 b.a=2", resolved)
}

func TestTemplateReferencesVariables(t *testing.T) {
	cases := []struct {
		tmpl any
		want bool
	}{
		{`SELECT 1`, false},
		{`{{ .args.x }}`, false},
		{`{{ .user.domain }}`, false},
		{`{{ .environment }}`, false},
		{`{{ range .args.ids }}{{ . }}{{ end }}`, false},
		{`{{ with .args }}{{ .x }}{{ end }}`, false},
		// Inside range and with, ".env" is a field of the rebound value, not the project variables.
		{`{{ with .args }}{{ .env }}{{ end }}`, false},
		{`{{ range .args.ids }}{{ .env.x }}{{ end }}`, false},
		{`{{ with .args }}{{ $.env.x }}{{ end }}`, true},
		{`{{ with .env }}{{ .x }}{{ end }}`, true},
		{`{{ metrics_sql "SELECT 1" }}`, false},
		// A nested template inside a string literal is resolved later by the template function, so it counts.
		{`{{ metrics_sql "SELECT {{ .env.x }}" }}`, true},
		{`{{ metrics_sql "SELECT 1 FROM mv" }}`, false},
		// Parse errors surface when the template is resolved.
		{`{{ .env`, false},
		{`{{ .env.x }}`, true},
		{`{{ .vars.x }}`, true},
		{`{{ .env }}`, true},
		{`{{ index .env "x" }}`, true},
		{`{{ env "x" }}`, true},
		{`{{ toJson .env }}`, true},
		{`{{ .env.x | b64enc }}`, true},
		{`{{ if .env.x }}a{{ end }}`, true},
		{`{{ if .args.x }}a{{ else }}{{ .env.x }}{{ end }}`, true},
		{`{{ range .args.ids }}{{ $.env.x }}{{ end }}`, true},
		{`{{ range .env }}{{ . }}{{ end }}`, true},
		{`{{ with .args }}{{ else }}{{ . }}{{ end }}`, true},
		{`{{ $e := .env }}{{ $e.x }}`, true},
		{`{{ . }}`, true},
		{`{{ $ }}`, true},
		{`{{ (index . "env").x }}`, true},
		// Any use of the full template data is rejected, even when it never reaches the variables, since the walker does not follow where it goes.
		{`{{ index . "user" "email" }}`, true},
		{`{{ $root := . }}{{ $root.user.email }}`, true},
		{`{{ with $ }}{{ .user.email }}{{ end }}`, true},
		{`{{ define "item" }}- {{ . }}{{ end }}{{ range .args.items }}{{ template "item" . }}{{ end }}`, true},
		{`{{ printf "%v" . }}`, true},
		{`{{ define "t" }}{{ .env.x }}{{ end }}{{ template "t" .args }}`, true},
		{`{{ template "t" . }}`, true},
		{map[string]any{"path": "s3://bucket/{{ .env.prefix }}/*.parquet"}, true},
		{map[string]any{"path": "s3://bucket/*.parquet", "partition": "{{ .args.p }}"}, false},
		{map[string]any{"list": []any{"a", "{{ .vars.x }}"}}, true},
		{[]any{1, "{{ .env.x }}"}, true},
		{nil, false},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.tmpl), func(t *testing.T) {
			require.Equal(t, tc.want, TemplateReferencesVariables(tc.tmpl))
		})
	}
}

func TestAsSQLList(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name:     "strings",
			template: `{{ as_sql_list (list "a" "b" "c") }}`,
			want:     "('a', 'b', 'c')",
		},
		{
			name:     "ints",
			template: `{{ as_sql_list (list 1 2 3) }}`,
			want:     "(1, 2, 3)",
		},
		{
			name:     "mixed types",
			template: `{{ as_sql_list (list 1 "b" 3) }}`,
			want:     "(1, 'b', 3)",
		},
		{
			name:     "empty list",
			template: `{{ as_sql_list (list) }}`,
			want:     "()",
		},
		{
			name:     "nil input",
			template: `{{ as_sql_list nil }}`,
			want:     "()",
		},
		{
			name:     "single string",
			template: `{{ as_sql_list "hello" }}`,
			want:     "('hello')",
		},
		{
			name:     "string with quotes",
			template: `{{ as_sql_list "hello'world" }}`,
			want:     "('hello''world')",
		},
		{
			name:     "booleans",
			template: `{{ as_sql_list (list true false) }}`,
			want:     "(true, false)",
		},
		{
			name:     "floats",
			template: `{{ as_sql_list (list 1.5 2.7 3.14) }}`,
			want:     "(1.5, 2.7, 3.14)",
		},
		{
			name:     "mixed with null",
			template: `{{ as_sql_list (list "a" nil "c") }}`,
			want:     "('a', NULL, 'c')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveTemplate(tt.template, TemplateData{}, false)
			require.NoError(t, err)
			require.Equal(t, tt.want, resolved)
		})
	}
}

func TestAsSQLListSecurityPolicies(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     TemplateData
		want     string
	}{
		{
			name:     "single restaurant access",
			template: `restaurant_id = {{ .user.restaurant_id }}`,
			data: TemplateData{
				User: map[string]any{
					"restaurant_id": "rest_123",
				},
			},
			want: "restaurant_id = rest_123",
		},
		{
			name:     "multi-valued campaigns with as_sql_list",
			template: `campaign IN {{ .user.allowed_campaigns | as_sql_list }}`,
			data: TemplateData{
				User: map[string]any{
					"allowed_campaigns": []string{"campaign1", "campaign2", "camp\"ign3"},
				},
			},
			want: `campaign IN ('campaign1', 'campaign2', 'camp"ign3')`,
		},
		{
			name:     "mixed types with admin flag and numeric budget",
			template: `region = '{{ .user.region }}' AND budget > {{ .user.min_budget }}`,
			data: TemplateData{
				User: map[string]any{
					"region":     "us-west",
					"min_budget": 1000,
				},
			},
			want: "region = 'us-west' AND budget > 1000",
		},
		{
			name:     "access check with len function",
			template: `{{ len .user.allowed_campaigns }} > 0`,
			data: TemplateData{
				User: map[string]any{
					"allowed_campaigns": []string{"campaign1", "campaign2"},
				},
			},
			want: "2 > 0",
		},
		{
			name:     "boolean admin check",
			template: `{{ .user.is_admin }} == true`,
			data: TemplateData{
				User: map[string]any{
					"is_admin": true,
				},
			},
			want: "true == true",
		},
		{
			name:     "null check for access control",
			template: `{{ .user.restaurant_id }} IS NOT NULL`,
			data: TemplateData{
				User: map[string]any{
					"restaurant_id": "rest_456",
				},
			},
			want: "rest_456 IS NOT NULL",
		},
		{
			name:     "complex security policy with quotes and mixed types",
			template: `user_id IN {{ .user.allowed_users | as_sql_list }} AND priority >= {{ .user.min_priority }} AND status = '{{ .user.status }}'`,
			data: TemplateData{
				User: map[string]any{
					"allowed_users": []interface{}{"user'1", "user2", 123},
					"min_priority":  5,
					"status":        "active",
				},
			},
			want: `user_id IN ('user''1', 'user2', 123) AND priority >= 5 AND status = 'active'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveTemplate(tt.template, tt.data, false)
			require.NoError(t, err)
			require.Equal(t, tt.want, resolved)
		})
	}
}
