package admin

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

// TriggerReport triggers an ad-hoc run of a report
func (s *Service) TriggerReport(ctx context.Context, depl *database.Deployment, report string) (err error) {
	names := []*runtimev1.ResourceName{
		{
			Kind: runtime.ResourceKindReport,
			Name: report,
		},
	}

	rt, err := s.openRuntimeClientForDeployment(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	_, err = rt.CreateTrigger(ctx, &runtimev1.CreateTriggerRequest{
		InstanceId: depl.RuntimeInstanceID,
		Trigger: &runtimev1.CreateTriggerRequest_RefreshTriggerSpec{
			RefreshTriggerSpec: &runtimev1.RefreshTriggerSpec{OnlyNames: names},
		},
	})
	return err
}

// TriggerReconcileAndAwaitReport triggers a reconcile and polls the runtime until the given report's spec version has been updated (or ctx is canceled).
func (s *Service) TriggerReconcileAndAwaitReport(ctx context.Context, depl *database.Deployment, reportName string) error {
	rt, err := s.openRuntimeClientForDeployment(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	reportReq := &runtimev1.GetResourceRequest{
		InstanceId: depl.RuntimeInstanceID,
		Name: &runtimev1.ResourceName{
			Kind: runtime.ResourceKindReport,
			Name: reportName,
		},
	}

	// Get old spec version
	var oldSpecVersion *int64
	r, err := rt.GetResource(ctx, reportReq)
	if err == nil {
		oldSpecVersion = &r.Resource.Meta.SpecVersion
	}

	// Trigger reconcile
	_, err = rt.CreateTrigger(ctx, &runtimev1.CreateTriggerRequest{
		InstanceId: depl.RuntimeInstanceID,
		Trigger: &runtimev1.CreateTriggerRequest_PullTriggerSpec{
			PullTriggerSpec: &runtimev1.PullTriggerSpec{},
		},
	})
	if err != nil {
		return err
	}

	// Poll every 1 seconds until the report is found or the ctx is cancelled
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		r, err := rt.GetResource(ctx, reportReq)
		if err != nil {
			if oldSpecVersion != nil {
				// Success - previously the report was found, now we cannot find it anymore
				return nil
			}
			// Continue polling
			continue
		}
		if oldSpecVersion == nil {
			// Success - previously the report was not found, now we found one
			return nil
		}
		if *oldSpecVersion != r.Resource.Meta.SpecVersion {
			// Success - the spec version has changed
			return nil
		}
	}
}

// LookupReport fetches a report's spec from a runtime deployment.
func (s *Service) LookupReport(ctx context.Context, depl *database.Deployment, reportName string) (*runtimev1.ReportSpec, error) {
	rt, err := s.openRuntimeClientForDeployment(depl)
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
