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
		mv   *runtimev1.MetricsView
	}
	tests := []struct {
		name           string
		args           args
		want           *ResolvedMetricsViewPolicy
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
				mv: &runtimev1.MetricsView{
					Name: "test_domain",
					Policy: &runtimev1.MetricsView_Policy{
						HasAccess: "{{.user.admin}}",
						Filter:    "WHERE domain = '{{.user.domain}}'",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewPolicy{
				HasAccess: true,
				Filter:    "WHERE domain = 'rilldata.com'",
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
				mv: &runtimev1.MetricsView{
					Name: "test_group",
					Policy: &runtimev1.MetricsView_Policy{
						HasAccess: "'{{.user.domain}}' = 'rilldata.com'",
						Filter:    "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewPolicy{
				HasAccess: true,
				Filter:    "WHERE groups IN ('test')",
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
				mv: &runtimev1.MetricsView{
					Name: "test_groups",
					Policy: &runtimev1.MetricsView_Policy{
						HasAccess: "'{{.user.domain}}' = 'rilldata.com'",
						Filter:    "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewPolicy{
				HasAccess: true,
				Filter:    "WHERE groups IN ('g1', 'g2')",
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
				mv: &runtimev1.MetricsView{
					Name: "test_no_groups",
					Policy: &runtimev1.MetricsView_Policy{
						HasAccess: "{{.user.admin}}",
						Filter:    "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewPolicy{
				HasAccess: false,
				Filter:    "WHERE groups IN ('')",
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
				mv: &runtimev1.MetricsView{
					Name: "test_include",
					Policy: &runtimev1.MetricsView_Policy{
						HasAccess: "'{{.user.domain}}' = 'rilldata.com'",
						Filter:    "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include: []*runtimev1.MetricsView_Policy_FieldCondition{
							{
								Name:      "col1",
								Condition: "'{{.user.domain}}' = 'test.com'",
							},
							{
								Name:      "col2",
								Condition: "'{{.user.domain}}' = 'rilldata.com'",
							},
							{
								Name:      "col3",
								Condition: "{{.user.admin}}",
							},
						},
						Exclude: nil,
					},
				},
			},
			want: &ResolvedMetricsViewPolicy{
				HasAccess: true,
				Filter:    "WHERE groups IN ('all')",
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
				mv: &runtimev1.MetricsView{
					Name: "test_include_empty_condition",
					Policy: &runtimev1.MetricsView_Policy{
						HasAccess: "'{{.user.domain}}' = 'rilldata.com'",
						Filter:    "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include: []*runtimev1.MetricsView_Policy_FieldCondition{
							{
								Name:      "col1",
								Condition: "",
							},
							{
								Name:      "col2",
								Condition: "'{{.user.domain}}' = 'rilldata.com'",
							},
						},
						Exclude: nil,
					},
				},
			},
			want: &ResolvedMetricsViewPolicy{
				HasAccess: false,
				Filter:    "",
				Include:   []string{},
				Exclude:   nil,
			},
			wantErr:        true,
			errMsgContains: "cannot evaluate empty expression",
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
				mv: &runtimev1.MetricsView{
					Name: "test_include_empty_condition",
					Policy: &runtimev1.MetricsView_Policy{
						HasAccess: "'{{.user.domain}}' = 'rilldata.com'",
						Filter:    "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude: []*runtimev1.MetricsView_Policy_FieldCondition{
							{
								Name:      "col1",
								Condition: "'{{.user.domain}}' = 'test.com'",
							},
							{
								Name:      "col2",
								Condition: "'{{.user.domain}}' = 'rilldata.com'",
							},
						},
					},
				},
			},
			want: &ResolvedMetricsViewPolicy{
				HasAccess: true,
				Filter:    "WHERE groups IN ('all')",
				Include:   nil,
				Exclude:   []string{"col2"},
			},
			wantErr: false,
		},
		{
			name: "test_empty_policy",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsView{
					Name:   "test_empty_policy",
					Policy: &runtimev1.MetricsView_Policy{},
				},
			},
			want: &ResolvedMetricsViewPolicy{
				HasAccess: false,
				Filter:    "",
				Include:   nil,
				Exclude:   nil,
			},
			wantErr: false,
		},
		{
			name: "test_nil_policy",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsView{
					Name:   "test_nil_policy",
					Policy: nil,
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "test_empty_user_attr",
			args: args{
				attr: nil,
				mv: &runtimev1.MetricsView{
					Name: "test",
					Policy: &runtimev1.MetricsView_Policy{
						HasAccess: "'{{.user.domain}}' = 'rilldata.com'",
						Filter:    "WHERE domain = '{{.user.domain}}'",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			// since aud is nil in test case, open policy will be applied which same as local dev experience
			want: &ResolvedMetricsViewPolicy{
				HasAccess: true,
				Filter:    "",
				Include:   nil,
				Exclude:   nil,
			},
			wantErr: false,
		},
		{
			name: "test_empty_has_access",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []interface{}{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsView{
					Name: "test_empty_has_access",
					Policy: &runtimev1.MetricsView_Policy{
						Filter:  "WHERE domain = '{{.user.domain}}'",
						Include: nil,
						Exclude: nil,
					},
				},
			},
			want: &ResolvedMetricsViewPolicy{
				HasAccess: false,
				Filter:    "WHERE domain = 'rilldata.com'",
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
				mv: &runtimev1.MetricsView{
					Name: "test",
					Policy: &runtimev1.MetricsView_Policy{
						HasAccess: "'{{.user.domain}}' = 'rilldata.com' OR '{{.user.domain}}' = 'gmail.com'",
						Filter:    "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewPolicy{
				HasAccess: false,
				Filter:    "WHERE groups IN ('test')",
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
				mv: &runtimev1.MetricsView{
					Name: "test",
					Policy: &runtimev1.MetricsView_Policy{
						HasAccess: "('{{.user.domain}}' = 'rilldata.com' OR '{{.user.domain}}' = 'gmail.com') AND {{.user.admin}}",
						Filter:    "WHERE groups IN ('{{ .user.groups | join \"', '\" }}')",
						Include:   nil,
						Exclude:   nil,
					},
				},
			},
			want: &ResolvedMetricsViewPolicy{
				HasAccess: true,
				Filter:    "WHERE groups IN ('test')",
				Include:   nil,
				Exclude:   nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPolicyEngine(1, zap.NewNop())
			got, err := p.resolveMetricsViewPolicy(tt.args.attr, "", tt.args.mv, time.Now())
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsgContains) {
					t.Errorf("ResolveMetricsViewPolicy() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ResolveMetricsViewPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolveMetricsViewPolicy() got = %v, want %v", got, tt.want)
			}
		})
	}
}
