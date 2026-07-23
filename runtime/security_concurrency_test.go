package runtime

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestResolveSecurityDoesNotMutateSharedClaims(t *testing.T) {
	// Magic-token claims are reused across requests. Built-in recipient access
	// must extend a request-local rule rather than mutate that shared token state.
	claims := &SecurityClaims{
		UserAttributes: map[string]any{"email": "recipient@example.com"},
		AdditionalRules: []*runtimev1.SecurityRule{{
			Rule: &runtimev1.SecurityRule_Access{Access: &runtimev1.SecurityRuleAccess{
				Allow:          true,
				Exclusive:      true,
				ConditionKinds: []string{ResourceKindMetricsView},
			}},
		}},
	}
	before, err := json.Marshal(claims)
	require.NoError(t, err)

	const resolutions = 64
	engine := newSecurityEngine(resolutions*2, zap.NewNop(), nil)
	results := make([]bool, resolutions)
	errs := make([]error, resolutions)
	var wg sync.WaitGroup
	wg.Add(resolutions)
	for i := 0; i < resolutions; i++ {
		i := i
		go func() {
			defer wg.Done()
			kind := ResourceKindReport
			if i%2 == 1 {
				kind = ResourceKindAlert
			}
			res := newSecurityTestNotificationResource(kind, fmt.Sprintf("notification-%d", i), "recipient@example.com")
			security, resolveErr := engine.resolveSecurity(t.Context(), "instance", "prod", nil, claims, res)
			errs[i] = resolveErr
			if resolveErr == nil {
				results[i] = security.CanAccess()
			}
		}()
	}
	wg.Wait()

	for i := range errs {
		require.NoErrorf(t, errs[i], "resolution %d", i)
		require.Truef(t, results[i], "recipient should access notification %d", i)
	}
	after, err := json.Marshal(claims)
	require.NoError(t, err)
	require.JSONEq(t, string(before), string(after))
	require.Empty(t, claims.AdditionalRules[0].GetAccess().ConditionResources)
}

func newSecurityTestNotificationResource(kind, name, recipient string) *runtimev1.Resource {
	properties, err := structpb.NewStruct(map[string]any{"recipients": []any{recipient}})
	if err != nil {
		panic(err)
	}
	meta := &runtimev1.ResourceMeta{
		Name:           &runtimev1.ResourceName{Kind: kind, Name: name},
		StateUpdatedOn: timestamppb.New(time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)),
	}
	notifiers := []*runtimev1.Notifier{{Connector: "email", Properties: properties}}
	if kind == ResourceKindAlert {
		return &runtimev1.Resource{
			Meta: meta,
			Resource: &runtimev1.Resource_Alert{Alert: &runtimev1.Alert{
				Spec: &runtimev1.AlertSpec{Notifiers: notifiers},
			}},
		}
	}
	return &runtimev1.Resource{
		Meta: meta,
		Resource: &runtimev1.Resource_Report{Report: &runtimev1.Report{
			Spec: &runtimev1.ReportSpec{Notifiers: notifiers},
		}},
	}
}
