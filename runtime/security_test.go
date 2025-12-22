package runtime

import (
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestResolveMetricsView(t *testing.T) {
	type args struct {
		attr map[string]any
		mv   *runtimev1.MetricsViewSpec
	}
	tests := []struct {
		name            string
		args            args
		wantAccess      bool
		wantRowFilter   string
		wantFieldAccess map[string]bool
		wantErr         bool
		errMsgContains  string
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
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: "{{.user.admin}}", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "domain = '{{.user.domain}}'"}}},
					},
				},
			},
			wantAccess:    true,
			wantRowFilter: "domain = 'rilldata.com'",
			wantErr:       false,
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
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: "'{{.user.domain}}' = 'rilldata.com'", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
					},
				},
			},
			wantAccess:    true,
			wantRowFilter: "groups IN ('test')",
			wantErr:       false,
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
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: "'{{.user.domain}}' = 'rilldata.com'", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
					},
				},
			},
			wantAccess:    true,
			wantRowFilter: "groups IN ('g1', 'g2')",
			wantErr:       false,
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
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: "{{.user.admin}}", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
					},
				},
			},
			wantAccess:    false,
			wantRowFilter: "groups IN ('')",
			wantErr:       false,
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
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: "'{{.user.domain}}' = 'rilldata.com'", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "'{{.user.domain}}' = 'test.com'",
							Allow:               true,
							Fields:              []string{"col1"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:               true,
							Fields:              []string{"col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "{{.user.admin}}",
							Allow:               true,
							Fields:              []string{"col3"},
						}}},
					},
				},
			},
			wantAccess:      true,
			wantFieldAccess: map[string]bool{"col2": true, "col3": true},
			wantRowFilter:   "groups IN ('all')",
			wantErr:         false,
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
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: "'{{.user.domain}}' = 'rilldata.com'", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "'{{.user.domain}}' = 'test.com'",
							Allow:               true,
							Fields:              []string{"col1", "col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:               true,
							Fields:              []string{"col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "{{.user.admin}}",
							Allow:               true,
							Fields:              []string{"col3"},
						}}},
					},
				},
			},
			wantAccess:      true,
			wantFieldAccess: map[string]bool{"col2": true, "col3": true},
			wantRowFilter:   "groups IN ('all')",
			wantErr:         false,
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
							ConditionExpression: "",
							Allow:               true,
							Fields:              []string{"col1"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:               true,
							Fields:              []string{"col2"},
						}}},
					},
				},
			},
			wantAccess:      false,
			wantFieldAccess: map[string]bool{"col1": true, "col2": true},
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
							ConditionExpression: "'{{.user.domain}}' = 'test.com'",
							Allow:               true,
							Fields:              []string{"col1", "col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:               true,
							Fields:              []string{"col2"},
						}}},
					},
				},
			},
			wantFieldAccess: map[string]bool{},
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
							ConditionExpression: "'{{.user.domain}}' = 'test.com'",
							Allow:               false,
							Fields:              []string{"col1", "col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:               false,
							Fields:              []string{"col2"},
						}}},
					},
				},
			},
			wantFieldAccess: map[string]bool{"col2": false},
			wantErr:         false,
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
							ConditionExpression: "'{{.user.domain}}' = 'test.com'",
							Allow:               false,
							Fields:              []string{"col1", "col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "'{{.user.domain}}' != 'rilldata.com'",
							Allow:               false,
							Fields:              []string{"col3"},
						}}},
					},
				},
			},
			wantFieldAccess: map[string]bool{},
			wantErr:         false,
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
					Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
						{Name: "col1"},
						{Name: "col2"},
					},
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Allow:     true,
							AllFields: true,
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "'{{.user.domain}}' = 'test.com'",
							Allow:               false,
							Fields:              []string{"col1", "col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:               false,
							Fields:              []string{"col2"},
						}}},
					},
				},
			},
			wantFieldAccess: map[string]bool{"col1": true, "col2": false},
			wantErr:         false,
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
					Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
						{Name: "col1"},
						{Name: "col2"},
					},
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							Allow:     true,
							AllFields: true,
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "'{{.user.domain}}' = 'test.com'",
							Allow:               false,
							Fields:              []string{"col1", "col2"},
						}}},
						{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{
							ConditionExpression: "'{{.user.domain}}' = 'rilldata.com'",
							Allow:               false,
							Fields:              []string{"col2"},
						}}},
					},
				},
			},
			wantFieldAccess: map[string]bool{"col1": true, "col2": true},
			wantErr:         false,
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
			wantAccess: true,
			wantErr:    false,
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
			wantAccess: true,
			wantErr:    false,
		},
		{
			name: "test_empty_user_attr",
			args: args{
				attr: nil,
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: "'{{.user.domain}}' = 'rilldata.com'", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "domain = '{{.user.domain}}'"}}},
					},
				},
			},
			wantAccess:    false,
			wantRowFilter: "domain = '<no value>'",
			wantErr:       false,
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
			wantAccess:    false,
			wantRowFilter: "domain = 'rilldata.com'",
			wantErr:       false,
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
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: "'{{.user.domain}}' = 'rilldata.com' OR '{{.user.domain}}' = 'gmail.com'", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
					},
				},
			},
			wantAccess:    false,
			wantRowFilter: "groups IN ('test')",
			wantErr:       false,
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
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: "('{{.user.domain}}' = 'rilldata.com' OR '{{.user.domain}}' = 'gmail.com') AND {{.user.admin}}", Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "groups IN ('{{ .user.groups | join \"', '\" }}')"}}},
					},
				},
			},
			wantAccess:    true,
			wantRowFilter: "groups IN ('test')",
			wantErr:       false,
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
					MetricsView: &runtimev1.MetricsView{
						Spec: tt.args.mv,
						State: &runtimev1.MetricsViewState{
							ValidSpec: tt.args.mv,
						},
					},
				},
			}

			claims := &SecurityClaims{UserAttributes: tt.args.attr}
			p := newSecurityEngine(1, zap.NewNop(), nil)
			got, err := p.resolveSecurity(t.Context(), "", "test", map[string]string{}, claims, r)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsgContains) {
					t.Errorf("ResolveSecurity() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ResolveSecurity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.Equal(t, tt.wantAccess, got.CanAccess())
			require.Equal(t, tt.wantRowFilter, got.RowFilter())
			if tt.wantFieldAccess != nil || got != nil {
				require.Equal(t, tt.wantFieldAccess, got.fieldAccess)
			}
		})
	}
}
