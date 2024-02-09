package admin

import (
	"context"

	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

// LookupAlert fetches a alert's spec from a runtime deployment.
func (s *Service) LookupAlert(ctx context.Context, depl *database.Deployment, alertName string) (*runtimev1.AlertSpec, error) {
	rt, err := s.openRuntimeClientForDeployment(depl)
	if err != nil {
		return nil, err
	}
	defer rt.Close()

	res, err := rt.GetResource(ctx, &runtimev1.GetResourceRequest{
		InstanceId: depl.RuntimeInstanceID,
		Name: &runtimev1.ResourceName{
			Kind: runtime.ResourceKindAlert,
			Name: alertName,
		},
	})
	if err != nil {
		return nil, err
	}

	return res.Resource.Resource.(*runtimev1.Resource_Alert).Alert.Spec, nil
}
