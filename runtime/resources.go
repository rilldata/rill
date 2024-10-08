package runtime

import (
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	compilerv1 "github.com/rilldata/rill/runtime/compilers/rillv1"
)

// Built-in resource kinds
const (
	ResourceKindProjectParser  string = "rill.runtime.v1.ProjectParser"
	ResourceKindSource         string = "rill.runtime.v1.Source"
	ResourceKindModel          string = "rill.runtime.v1.Model"
	ResourceKindMetricsView    string = "rill.runtime.v1.MetricsView"
	ResourceKindExplore        string = "rill.runtime.v1.Explore"
	ResourceKindMigration      string = "rill.runtime.v1.Migration"
	ResourceKindReport         string = "rill.runtime.v1.Report"
	ResourceKindAlert          string = "rill.runtime.v1.Alert"
	ResourceKindPullTrigger    string = "rill.runtime.v1.PullTrigger"
	ResourceKindRefreshTrigger string = "rill.runtime.v1.RefreshTrigger"
	ResourceKindBucketPlanner  string = "rill.runtime.v1.BucketPlanner"
	ResourceKindTheme          string = "rill.runtime.v1.Theme"
	ResourceKindComponent      string = "rill.runtime.v1.Component"
	ResourceKindCanvas         string = "rill.runtime.v1.Canvas"
	ResourceKindAPI            string = "rill.runtime.v1.API"
	ResourceKindConnector      string = "rill.runtime.v1.Connector"
)

// ResourceNameFromCompiler converts a compiler resource name to a runtime resource name.
func ResourceNameFromCompiler(name compilerv1.ResourceName) *runtimev1.ResourceName {
	switch name.Kind {
	case compilerv1.ResourceKindSource:
		return &runtimev1.ResourceName{Kind: ResourceKindSource, Name: name.Name}
	case compilerv1.ResourceKindModel:
		return &runtimev1.ResourceName{Kind: ResourceKindModel, Name: name.Name}
	case compilerv1.ResourceKindMetricsView:
		return &runtimev1.ResourceName{Kind: ResourceKindMetricsView, Name: name.Name}
	case compilerv1.ResourceKindExplore:
		return &runtimev1.ResourceName{Kind: ResourceKindExplore, Name: name.Name}
	case compilerv1.ResourceKindMigration:
		return &runtimev1.ResourceName{Kind: ResourceKindMigration, Name: name.Name}
	case compilerv1.ResourceKindReport:
		return &runtimev1.ResourceName{Kind: ResourceKindReport, Name: name.Name}
	case compilerv1.ResourceKindAlert:
		return &runtimev1.ResourceName{Kind: ResourceKindAlert, Name: name.Name}
	case compilerv1.ResourceKindTheme:
		return &runtimev1.ResourceName{Kind: ResourceKindTheme, Name: name.Name}
	case compilerv1.ResourceKindComponent:
		return &runtimev1.ResourceName{Kind: ResourceKindComponent, Name: name.Name}
	case compilerv1.ResourceKindCanvas:
		return &runtimev1.ResourceName{Kind: ResourceKindCanvas, Name: name.Name}
	case compilerv1.ResourceKindAPI:
		return &runtimev1.ResourceName{Kind: ResourceKindAPI, Name: name.Name}
	case compilerv1.ResourceKindConnector:
		return &runtimev1.ResourceName{Kind: ResourceKindConnector, Name: name.Name}
	default:
		panic(fmt.Errorf("unknown resource type %q", name.Kind))
	}
}

// ResourceNameToCompiler converts a runtime resource name to a compiler resource name.
func ResourceNameToCompiler(name *runtimev1.ResourceName) compilerv1.ResourceName {
	switch name.Kind {
	case ResourceKindSource:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindSource, Name: name.Name}
	case ResourceKindModel:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindModel, Name: name.Name}
	case ResourceKindMetricsView:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindMetricsView, Name: name.Name}
	case ResourceKindExplore:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindExplore, Name: name.Name}
	case ResourceKindMigration:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindMigration, Name: name.Name}
	case ResourceKindReport:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindReport, Name: name.Name}
	case ResourceKindAlert:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindAlert, Name: name.Name}
	case ResourceKindTheme:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindTheme, Name: name.Name}
	case ResourceKindComponent:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindComponent, Name: name.Name}
	case ResourceKindCanvas:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindCanvas, Name: name.Name}
	case ResourceKindAPI:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindAPI, Name: name.Name}
	case ResourceKindConnector:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindConnector, Name: name.Name}
	default:
		panic(fmt.Errorf("unknown resource type %q", name.Kind))
	}
}

// PrettifyResourceKind returns the resource kind in a user-friendly format suitable for printing.
func PrettifyResourceKind(k string) string {
	k = strings.TrimPrefix(k, "rill.runtime.v1.")
	k = strings.TrimSuffix(k, "V2")
	return k
}

// PrettifyReconcileStatus returns the reconcile status in a user-friendly format suitable for printing.
func PrettifyReconcileStatus(s runtimev1.ReconcileStatus) string {
	switch s {
	case runtimev1.ReconcileStatus_RECONCILE_STATUS_UNSPECIFIED:
		return "Unknown"
	case runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE:
		return "Idle"
	case runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING:
		return "Pending"
	case runtimev1.ReconcileStatus_RECONCILE_STATUS_RUNNING:
		return "Running"
	default:
		panic(fmt.Errorf("unknown reconcile status: %s", s.String()))
	}
}
