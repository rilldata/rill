package runtime

import (
	"reflect"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
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
					"groups": []any{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Condition: "{{.user.admin}}", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "domain = '{{.user.domain}}'"}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "domain = 'rilldata.com'",
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
					"groups": []any{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Condition: "'{{.user.domain}}' = 'rilldata.com'", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "groups IN ('test')",
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
					"groups": []any{"g1", "g2"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Condition: "'{{.user.domain}}' = 'rilldata.com'", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "groups IN ('g1', 'g2')",
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
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Condition: "{{.user.admin}}", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    false,
				RowFilter: "groups IN ('')",
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
					"groups": []any{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Condition: "'{{.user.domain}}' = 'rilldata.com'", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'test.com'",
							Allow:     true,
							Fields:    []string{"col1"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:     true,
							Fields:    []string{"col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "{{.user.admin}}",
							Allow:     true,
							Fields:    []string{"col3"},
						}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:      true,
				FieldAccess: map[string]bool{"col2": true, "col3": true},
				RowFilter:   "groups IN ('all')",
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
					"groups": []any{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Condition: "'{{.user.domain}}' = 'rilldata.com'", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'test.com'",
							Allow:     true,
							Fields:    []string{"col1", "col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:     true,
							Fields:    []string{"col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "{{.user.admin}}",
							Allow:     true,
							Fields:    []string{"col3"},
						}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:      true,
				FieldAccess: map[string]bool{"col2": true, "col3": true},
				RowFilter:   "groups IN ('all')",
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
					"groups": []any{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "",
							Allow:     true,
							Fields:    []string{"col1"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:     true,
							Fields:    []string{"col2"},
						}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:      false,
				FieldAccess: map[string]bool{"col1": true, "col2": true},
			},
		},
		{
			name: "test_no_include_matched",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@gmail.com",
					"domain": "gmail.com",
					"groups": []any{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'test.com'",
							Allow:     true,
							Fields:    []string{"col1", "col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:     true,
							Fields:    []string{"col2"},
						}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				FieldAccess: map[string]bool{},
			},
		},
		{
			name: "test_exclude",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []any{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'test.com'",
							Allow:     false,
							Fields:    []string{"col1", "col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:     false,
							Fields:    []string{"col2"},
						}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				FieldAccess: map[string]bool{"col2": false},
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
					"groups": []any{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'test.com'",
							Allow:     false,
							Fields:    []string{"col1", "col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' != 'rilldata.com'",
							Allow:     false,
							Fields:    []string{"col3"},
						}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				FieldAccess: map[string]bool{},
			},
			wantErr: false,
		},
		{
			name: "test_deny_precedence",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []any{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
						{Name: "col1"},
						{Name: "col2"},
					},
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Allow:     true,
							AllFields: true,
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'test.com'",
							Allow:     false,
							Fields:    []string{"col1", "col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:     false,
							Fields:    []string{"col2"},
						}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				FieldAccess: map[string]bool{"col1": true, "col2": false},
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
					"groups": []any{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
						{Name: "col1"},
						{Name: "col2"},
					},
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Allow:     true,
							AllFields: true,
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'test.com'",
							Allow:     false,
							Fields:    []string{"col1", "col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Condition: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:     false,
							Fields:    []string{"col2"},
						}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				FieldAccess: map[string]bool{"col1": true, "col2": true},
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
					"groups": []any{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "test_nil_Security",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
					"groups": []any{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "test_empty_user_attr",
			args: args{
				attr: nil,
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Condition: "'{{.user.domain}}' = 'rilldata.com'", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "domain = '{{.user.domain}}'"}}},
					},
				},
			},
			// since aud is nil in test case, open policy will be applied which same as local dev experience
			want: &ResolvedMetricsViewSecurity{
				Access: true,
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
					"groups": []any{"all"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "domain = '{{.user.domain}}'"}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    false,
				RowFilter: "domain = 'rilldata.com'",
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
					"groups": []any{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Condition: "'{{.user.domain}}' = 'rilldata.com' OR '{{.user.domain}}' = 'gmail.com'", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    false,
				RowFilter: "groups IN ('test')",
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
					"groups": []any{"test"},
					"admin":  true,
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Condition: "('{{.user.domain}}' = 'rilldata.com' OR '{{.user.domain}}' = 'gmail.com') AND {{.user.admin}}", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
					},
				},
			},
			want: &ResolvedMetricsViewSecurity{
				Access:    true,
				RowFilter: "groups IN ('test')",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &runtimev1.Resource{
				Meta: &runtimev1.ResourceMeta{
					Name: &runtimev1.ResourceName{
						Kind: ResourceKindMetricsView,
						Name: "test",
					},
					StateUpdatedOn: timestamppb.Now(),
				},
				Resource: &runtimev1.Resource_MetricsView{
					MetricsView: &runtimev1.MetricsViewV2{
						Spec: tt.args.mv,
						State: &runtimev1.MetricsViewState{
							ValidSpec: tt.args.mv,
						},
					},
				},
			}

			p := newSecurityEngine(1, zap.NewNop())
			got, err := p.resolveMetricsViewSecurity("", "test", tt.args.attr, nil, r)
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
