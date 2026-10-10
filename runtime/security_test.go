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
		{
			// test conditional row filter with custom "org_id" attribute
			name: "test_conditional_row_filter_custom_attribute",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@acme.com",
					"domain": "acme.com",
					"org_id": "acme",
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "{{ if .user.org_id }}deal_id IN (SELECT deal_id FROM deals WHERE org_id = '{{ .user.org_id }}'){{ else }}1=1{{ end }}"}}},
					},
				},
			},
			wantAccess:    true,
			wantRowFilter: "deal_id IN (SELECT deal_id FROM deals WHERE org_id = 'acme')",
			wantErr:       false,
		},
		{
			name: "test_conditional_row_filter_no_custom_attribute",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@rilldata.com",
					"domain": "rilldata.com",
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "{{ if .user.org_id }}deal_id IN (SELECT deal_id FROM deals WHERE org_id = '{{ .user.org_id }}'){{ else }}1=1{{ end }}"}}},
					},
				},
			},
			wantAccess:    true,
			wantRowFilter: "1=1",
			wantErr:       false,
		},
		{
			// An empty custom attribute is ignored, so the conditional row filter resolves to "1=1" (no restriction).
			name: "test_conditional_row_filter_empty_custom_attribute",
			args: args{
				attr: map[string]any{
					"name":   "test",
					"email":  "test@acme.com",
					"domain": "acme.com",
					"org_id": "",
				},
				mv: &runtimev1.MetricsViewSpec{
					SecurityRules: []*runtimev1.SecurityRule{
						{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Allow: true}}},
						{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "{{ if .user.org_id }}deal_id IN (SELECT deal_id FROM deals WHERE org_id = '{{ .user.org_id }}'){{ else }}1=1{{ end }}"}}},
					},
				},
			},
			wantAccess:    true,
			wantRowFilter: "1=1",
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

func TestResolveConnectorSecurity(t *testing.T) {
	adminAttrs := map[string]any{"email": "a@rilldata.com", "domain": "rilldata.com", "groups": []any{"eng"}, "admin": true}
	userAttrs := map[string]any{"email": "b@example.com", "domain": "example.com", "groups": []any{"sales"}, "admin": false}

	access := func(expr string) *runtimev1.SecurityRule {
		return &runtimev1.SecurityRule{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: expr, Allow: true}}}
	}
	fieldRule := func(expr string, allow bool, names ...string) *runtimev1.SecurityRule {
		return &runtimev1.SecurityRule{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{ConditionExpression: expr, Allow: allow, Fields: names}}}
	}
	allFields := func(allow bool) *runtimev1.SecurityRule {
		return &runtimev1.SecurityRule{Rule: &runtimev1.SecurityRule_FieldAccess{FieldAccess: &runtimev1.SecurityRuleFieldAccess{Allow: allow, AllFields: true}}}
	}

	tests := []struct {
		name       string
		rules      []*runtimev1.SecurityRule
		attrs      map[string]any
		stateTools []string
		tools      []string
		wantAccess bool
		wantTools  map[string]bool
	}{
		{
			name:       "no rules allows everyone and every tool",
			attrs:      userAttrs,
			tools:      []string{"search", "delete"},
			wantAccess: true,
			wantTools:  map[string]bool{"search": true, "delete": true, "unknown": true},
		},
		{
			name:       "access denied for non-admin",
			rules:      []*runtimev1.SecurityRule{access("{{ .user.admin }}")},
			attrs:      userAttrs,
			tools:      []string{"search"},
			wantAccess: false,
			wantTools:  map[string]bool{"search": false},
		},
		{
			name:       "access allowed for admin",
			rules:      []*runtimev1.SecurityRule{access("{{ .user.admin }}")},
			attrs:      adminAttrs,
			tools:      []string{"search"},
			wantAccess: true,
			wantTools:  map[string]bool{"search": true},
		},
		{
			name:       "include implies deny for unlisted tools",
			rules:      []*runtimev1.SecurityRule{access("true"), fieldRule("", true, "search")},
			attrs:      userAttrs,
			tools:      []string{"search", "delete"},
			wantAccess: true,
			wantTools:  map[string]bool{"search": true, "delete": false},
		},
		{
			name:       "conditional include for admins only",
			rules:      []*runtimev1.SecurityRule{access("true"), fieldRule("", true, "search"), fieldRule("{{ .user.admin }}", true, "create")},
			attrs:      userAttrs,
			tools:      []string{"search", "create"},
			wantAccess: true,
			wantTools:  map[string]bool{"search": true, "create": false},
		},
		{
			name:       "exclude wins over allow all",
			rules:      []*runtimev1.SecurityRule{access("true"), allFields(true), fieldRule("'{{ .user.domain }}' != 'rilldata.com'", false, "delete")},
			attrs:      userAttrs,
			tools:      []string{"search", "delete"},
			wantAccess: true,
			wantTools:  map[string]bool{"search": true, "delete": false},
		},
		{
			name:       "exclude condition not met keeps the tool",
			rules:      []*runtimev1.SecurityRule{access("true"), allFields(true), fieldRule("'{{ .user.domain }}' != 'rilldata.com'", false, "delete")},
			attrs:      adminAttrs,
			tools:      []string{"search", "delete"},
			wantAccess: true,
			wantTools:  map[string]bool{"search": true, "delete": true},
		},
		{
			name:       "tools from the connector state are used when none are given",
			rules:      []*runtimev1.SecurityRule{access("true"), allFields(true), fieldRule("", false, "delete")},
			attrs:      userAttrs,
			stateTools: []string{"search", "delete"},
			tools:      nil,
			wantAccess: true,
			wantTools:  map[string]bool{"search": true, "delete": false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stateTools []*runtimev1.MCPTool
			for _, name := range tt.stateTools {
				stateTools = append(stateTools, &runtimev1.MCPTool{Name: name})
			}
			r := &runtimev1.Resource{
				Meta: &runtimev1.ResourceMeta{
					Name:           &runtimev1.ResourceName{Kind: ResourceKindConnector, Name: "ext"},
					StateUpdatedOn: timestamppb.Now(),
				},
				Resource: &runtimev1.Resource_Connector{
					Connector: &runtimev1.ConnectorV2{
						Spec:  &runtimev1.ConnectorSpec{Driver: "mcp", SecurityRules: tt.rules},
						State: &runtimev1.ConnectorState{McpTools: stateTools},
					},
				},
			}
			claims := &SecurityClaims{UserID: "u1", UserAttributes: tt.attrs}
			p := newSecurityEngine(10, zap.NewNop(), nil)
			got, err := p.resolveConnectorSecurity(t.Context(), "", "test", map[string]string{}, claims, r, tt.tools)
			require.NoError(t, err)
			require.Equal(t, tt.wantAccess, got.CanAccess())
			for tool, want := range tt.wantTools {
				require.Equal(t, want, got.CanAccessField(tool), "tool %q", tool)
			}

			// The regular resolution of connector resources stays admin-only (it guards the connector's configuration, not its tools)
			regular, err := p.resolveSecurity(t.Context(), "", "test", map[string]string{}, claims, r)
			require.NoError(t, err)
			require.Equal(t, tt.attrs["admin"] == true, regular.CanAccess())
		})
	}
}

func TestResolveAPI(t *testing.T) {
	tests := []struct {
		name       string
		attr       map[string]any
		rules      []*runtimev1.SecurityRule
		wantAccess bool
		wantErr    bool
	}{
		{
			name:       "no_rules_allows",
			attr:       map[string]any{"tier": "free"},
			wantAccess: true,
		},
		{
			name:       "rule_matches_allows",
			attr:       map[string]any{"tier": "enterprise"},
			rules:      []*runtimev1.SecurityRule{{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: `{{ eq .user.tier "enterprise" }}`, Allow: true}}}},
			wantAccess: true,
		},
		{
			name:       "rule_does_not_match_denies",
			attr:       map[string]any{"tier": "free"},
			rules:      []*runtimev1.SecurityRule{{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: `{{ eq .user.tier "enterprise" }}`, Allow: true}}}},
			wantAccess: false,
		},
		{
			name:       "missing_attribute_denies",
			attr:       map[string]any{},
			rules:      []*runtimev1.SecurityRule{{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: `{{ eq .user.tier "enterprise" }}`, Allow: true}}}},
			wantAccess: false,
		},
		{
			name:    "evaluation_error_fails",
			attr:    map[string]any{"admin": "notabool"},
			rules:   []*runtimev1.SecurityRule{{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{ConditionExpression: `{{ .user.admin }}`, Allow: true}}}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &runtimev1.Resource{
				Meta: &runtimev1.ResourceMeta{
					Name:           &runtimev1.ResourceName{Kind: ResourceKindAPI, Name: "test"},
					StateUpdatedOn: timestamppb.Now(),
				},
				Resource: &runtimev1.Resource_Api{
					Api: &runtimev1.API{Spec: &runtimev1.APISpec{SecurityRules: tt.rules}},
				},
			}

			claims := &SecurityClaims{UserAttributes: tt.attr}
			p := newSecurityEngine(1, zap.NewNop(), nil)
			got, err := p.resolveSecurity(t.Context(), "", "test", map[string]string{}, claims, r)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantAccess, got.CanAccess())
		})
	}
}
