package runtime

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestResolveSecurityCacheIncludesTemplateVariables(t *testing.T) {
	// Security templates can depend on project variables, so two requests that
	// differ only by variables must not share a cached row filter.
	res := newSecurityTestMetricsView("orders", []*runtimev1.SecurityRule{
		{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Allow: true}}},
		{Rule: &runtimev1.SecurityRule_RowFilter{RowFilter: &runtimev1.SecurityRuleRowFilter{Sql: "region = '{{ .env.region }}'"}}},
	})
	engine := newSecurityEngine(10, zap.NewNop(), nil)
	claims := &SecurityClaims{}

	east, err := engine.resolveSecurity(t.Context(), "instance", "prod", map[string]string{"region": "east"}, claims, res)
	require.NoError(t, err)
	require.Equal(t, "region = 'east'", east.RowFilter())

	west, err := engine.resolveSecurity(t.Context(), "instance", "prod", map[string]string{"region": "west"}, claims, res)
	require.NoError(t, err)
	require.Equal(t, "region = 'west'", west.RowFilter())
}

func TestResolveSecurityCacheSeparatesResourceKinds(t *testing.T) {
	// Resource names are unique only within a kind. An API and a metrics view
	// with the same name and timestamp may intentionally have opposite policies.
	updated := timestamppb.New(time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC))
	api := &runtimev1.Resource{
		Meta:     &runtimev1.ResourceMeta{Name: &runtimev1.ResourceName{Kind: ResourceKindAPI, Name: "shared"}, StateUpdatedOn: updated},
		Resource: &runtimev1.Resource_Api{Api: &runtimev1.API{Spec: &runtimev1.APISpec{}}},
	}
	metrics := newSecurityTestMetricsView("shared", []*runtimev1.SecurityRule{
		{Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{Allow: false}}},
	})
	metrics.Meta.StateUpdatedOn = updated

	engine := newSecurityEngine(10, zap.NewNop(), nil)
	claims := &SecurityClaims{}

	apiSecurity, err := engine.resolveSecurity(t.Context(), "instance", "prod", nil, claims, api)
	require.NoError(t, err)
	require.True(t, apiSecurity.CanAccess())

	metricsSecurity, err := engine.resolveSecurity(t.Context(), "instance", "prod", nil, claims, metrics)
	require.NoError(t, err)
	require.False(t, metricsSecurity.CanAccess())
}

func newSecurityTestMetricsView(name string, rules []*runtimev1.SecurityRule) *runtimev1.Resource {
	spec := &runtimev1.MetricsViewSpec{SecurityRules: rules}
	return &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:           &runtimev1.ResourceName{Kind: ResourceKindMetricsView, Name: name},
			StateUpdatedOn: timestamppb.New(time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)),
		},
		Resource: &runtimev1.Resource_MetricsView{MetricsView: &runtimev1.MetricsView{
			Spec:  spec,
			State: &runtimev1.MetricsViewState{ValidSpec: spec},
		}},
	}
}
