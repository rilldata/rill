package runtime

import (
	"reflect"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"go.uber.org/zap"
)

func TestResolveMetricsView(t *testing.T) {
	type args struct {
		attr map[string]any
		mv   *runtimev1.MetricsViewSpec
	}
	tests := []struct {
		name           string
		args           args
		want           *ResolvedMetricsViewSecurity
		wantErr        bool
		errMsgContains string
	}{
		{
			name: "test_domain",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "{{.user.admin}}",
						RowFilter: "WHERE domain = '{{.user.domain}}'",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "WHERE domain = 'rilldata.com'",
				Include:   nil,
				Exclude:   nil,
			},
			wantErr: false,
		},
		{
			name: "test_group",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "'{{.user.domain}}' = 'rilldata.com'",
						RowFilter: "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "WHERE groups IN ('test')",
				Include:   nil,
				Exclude:   nil,
			},
			wantErr: false,
		},
		{
			name: "test_groups",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"g1", "g2"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "'{{.user.domain}}' = 'rilldata.com'",
						RowFilter: "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "WHERE groups IN ('g1', 'g2')",
				Include:   nil,
				Exclude:   nil,
			},
			wantErr: false,
		},
		{
			name: "test_no_groups",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": nil,
					"admin":  false,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "{{.user.admin}}",
						RowFilter: "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    false,
				RowFilter: "WHERE groups IN ('')",
				Include:   nil,
				Exclude:   nil,
			},
			wantErr: false,
		},
		{
			name: "test_include",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "'{{.user.domain}}' = 'rilldata.com'",
						RowFilter: "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include: []*runtimev1.MetricsViewSpec_SecurityV2_FieldConditionV2{
							{
								Names:     []string{"col1"},
								Condition: "'{{.user.domain}}' = 'test.com'",
							},
							{
								Names:     []string{"col2"},
								Condition: "'{{.user.domain}}' = 'rilldata.com'",
							},
							{
								Names:     []string{"col3"},
								Condition: "{{.user.admin}}",
							},
						},
						Exclude: nil,
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "WHERE groups IN ('all')",
				Include:   []string{"col2", "col3"},
				Exclude:   nil,
			},
			wantErr: false,
		},
		{
			name: "test_include_list",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "'{{.user.domain}}' = 'rilldata.com'",
						RowFilter: "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include: []*runtimev1.MetricsViewSpec_SecurityV2_FieldConditionV2{
							{
								Names:     []string{"col1", "col2"},
								Condition: "'{{.user.domain}}' = 'test.com'",
							},
							{
								Names:     []string{"col2"},
								Condition: "'{{.user.domain}}' = 'rilldata.com'",
							},
							{
								Names:     []string{"col3"},
								Condition: "{{.user.admin}}",
							},
						},
						Exclude: nil,
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "WHERE groups IN ('all')",
				Include:   []string{"col2", "col3"},
				Exclude:   nil,
			},
			wantErr: false,
		},
		{
			name: "test_include_empty_condition",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "'{{.user.domain}}' = 'rilldata.com'",
						RowFilter: "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include: []*runtimev1.MetricsViewSpec_SecurityV2_FieldConditionV2{
							{
								Names:     []string{"col1"},
								Condition: "",
							},
							{
								Names:     []string{"col2"},
								Condition: "'{{.user.domain}}' = 'rilldata.com'",
							},
						},
						Exclude: nil,
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    false,
				RowFilter: "",
				Include:   []string{},
				Exclude:   nil,
			},
			wantErr:        true,
			errMsgContains: "cannot evaluate empty expression",
		},
		{
			name: "test_no_include_matched",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@gmail.com",
					"domain": "gmail.com",
					"groups": []interface{}{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "('{{.user.domain}}' = 'rilldata.com' OR '{{.user.domain}}' = 'gmail.com') AND {{.user.admin}}",
						RowFilter: "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include: []*runtimev1.MetricsViewSpec_SecurityV2_FieldConditionV2{
							{
								Names:     []string{"col1"},
								Condition: "'{{.user.domain}}' = 'test.com'",
							},
							{
								Names:     []string{"col2"},
								Condition: "'{{.user.domain}}' = 'rilldata.com'",
							},
						},
						Exclude: nil,
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:     true,
				RowFilter:  "WHERE groups IN ('test')",
				Include:    nil,
				Exclude:    nil,
				ExcludeAll: true,
			},
			wantErr: false,
		},
		{
			name: "test_exclude",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "'{{.user.domain}}' = 'rilldata.com'",
						RowFilter: "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude: []*runtimev1.MetricsViewSpec_SecurityV2_FieldConditionV2{
							{
								Names:     []string{"col1"},
								Condition: "'{{.user.domain}}' = 'test.com'",
							},
							{
								Names:     []string{"col2"},
								Condition: "'{{.user.domain}}' = 'rilldata.com'",
							},
						},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "WHERE groups IN ('all')",
				Include:   nil,
				Exclude:   []string{"col2"},
			},
			wantErr: false,
		},
		{
			name: "test_exclude_list",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "'{{.user.domain}}' = 'rilldata.com'",
						RowFilter: "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude: []*runtimev1.MetricsViewSpec_SecurityV2_FieldConditionV2{
							{
								Names:     []string{"col1", "col2"},
								Condition: "'{{.user.domain}}' = 'test.com'",
							},
							{
								Names:     []string{"col2"},
								Condition: "'{{.user.domain}}' = 'rilldata.com'",
							},
						},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "WHERE groups IN ('all')",
				Include:   nil,
				Exclude:   []string{"col2"},
			},
			wantErr: false,
		},
		{
			name: "test_no_exclude_matched",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@gmail.com",
					"domain": "gmail.com",
					"groups": []interface{}{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "('{{.user.domain}}' = 'rilldata.com' OR '{{.user.domain}}' = 'gmail.com') AND {{.user.admin}}",
						RowFilter: "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude: []*runtimev1.MetricsViewSpec_SecurityV2_FieldConditionV2{
							{
								Names:     []string{"col1", "col2"},
								Condition: "'{{.user.domain}}' = 'test.com'",
							},
							{
								Names:     []string{"col2"},
								Condition: "'{{.user.domain}}' = 'rilldata.com'",
							},
						},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:     true,
				RowFilter:  "WHERE groups IN ('test')",
				Include:    nil,
				Exclude:    nil,
				ExcludeAll: false,
			},
			wantErr: false,
		},
		{
			name: "test_empty_Security",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    false,
				RowFilter: "",
				Include:   nil,
				Exclude:   nil,
			},
			wantErr: false,
		},
		{
			name: "test_nil_Security",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: nil,
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "test_empty_user_attr",
			args: args{
				attr: nil,
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "'{{.user.domain}}' = 'rilldata.com'",
						RowFilter: "WHERE domain = '{{.user.domain}}'",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			// since aud is nil in test case, open policy will be applied which same as local dev experience
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "",
				Include:   nil,
				Exclude:   nil,
			},
			wantErr: false,
		},
		{
			name: "test_empty_access",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						RowFilter: "WHERE domain = '{{.user.domain}}'",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    false,
				RowFilter: "WHERE domain = 'rilldata.com'",
				Include:   nil,
				Exclude:   nil,
			},
			wantErr: false,
		},
		{
			name: "test_composite_condition_1",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@exclude.com",
					"domain": "exclude.com",
					"groups": []interface{}{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "'{{.user.domain}}' = 'rilldata.com' OR '{{.user.domain}}' = 'gmail.com'",
						RowFilter: "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    false,
				RowFilter: "WHERE groups IN ('test')",
				Include:   nil,
				Exclude:   nil,
			},
			wantErr: false,
		},
		{
			name: "test_composite_condition_2",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Security: &runtimev1.MetricsViewSpec_SecurityV2{
						Access:    "('{{.user.domain}}' = 'rilldata.com' OR '{{.user.domain}}' = 'gmail.com') AND {{.user.admin}}",
						RowFilter: "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "WHERE groups IN ('test')",
				Include:   nil,
				Exclude:   nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newSecurityEngine(1, zap.NewNop())
			got, err := p.resolveMetricsViewSecurity(tt.args.attr, "", tt.args.mv, time.Now())
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsgContains) {
					t.Errorf("ResolveMetricsViewSecurity() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ResolveMetricsViewSecurity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolveMetricsViewSecurity() got = %v, want %v", got, tt.want)
			}
		})
	}
}
