package admin

import (
	"context"

	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

// LookupCanvas fetches a canvas's spec from a runtime deployment.
func (s *Service) LookupCanvas(ctx context.Context, depl *database.Deployment, canvasName string) (*runtimev1.CanvasSpec, error) {
	rt, err := s.OpenRuntimeClient(depl)
	if err != nil {
		return nil, err
	}
	defer rt.Close()

	res, err := rt.GetResource(ctx, &runtimev1.GetResourceRequest{
		InstanceId: depl.RuntimeInstanceID,
		Name: &runtimev1.ResourceName{
			Kind: runtime.ResourceKindCanvas,
			Name: canvasName,
		},
	})
	if err != nil {
		return nil, err
	}

	return res.Resource.Resource.(*runtimev1.Resource_Canvas).Canvas.Spec, nil
}
