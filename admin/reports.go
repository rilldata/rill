package admin

import (
	"context"

	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

// TriggerReport triggers an ad-hoc run of a report
func (s *Service) TriggerReport(ctx context.Context, depl *database.Deployment, report string) (err error) {
	rt, err := s.OpenRuntimeClient(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	_, err = rt.CreateTrigger(ctx, &runtimev1.CreateTriggerRequest{
		InstanceId: depl.RuntimeInstanceID,
		Resources: []*runtimev1.ResourceName{
			{Kind: runtime.ResourceKindReport, Name: report},
		},
	})
	return err
}

// LookupReport fetches a report's spec from a runtime deployment.
func (s *Service) LookupReport(ctx context.Context, depl *database.Deployment, reportName string) (*runtimev1.ReportSpec, error) {
	rt, err := s.OpenRuntimeClient(depl)
	if err != nil {
		return nil, err
	}
	defer rt.Close()

	res, err := rt.GetResource(ctx, &runtimev1.GetResourceRequest{
		InstanceId: depl.RuntimeInstanceID,
		Name: &runtimev1.ResourceName{
			Kind: runtime.ResourceKindReport,
			Name: reportName,
		},
	})
	if err != nil {
		return nil, err
	}

	return res.Resource.Resource.(*runtimev1.Resource_Report).Report.Spec, nil
}
