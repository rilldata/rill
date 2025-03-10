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

// ResourceKindFromPretty converts a user-friendly resource kind to a runtime resource kind.
func ResourceKindFromShorthand(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "projectparser", "project_parser":
		return ResourceKindProjectParser
	case "source":
		return ResourceKindSource
	case "model":
		return ResourceKindModel
	case "metricsview", "metrics_view":
		return ResourceKindMetricsView
	case "explore":
		return ResourceKindExplore
	case "migration":
		return ResourceKindMigration
	case "report":
		return ResourceKindReport
	case "alert":
		return ResourceKindAlert
	case "pulltrigger", "pull_trigger":
		return ResourceKindPullTrigger
	case "refreshtrigger", "refresh_trigger":
		return ResourceKindRefreshTrigger
	case "bucketplanner", "bucket_planner":
		return ResourceKindBucketPlanner
	case "theme":
		return ResourceKindTheme
	case "component":
		return ResourceKindComponent
	case "canvas":
		return ResourceKindCanvas
	case "api":
		return ResourceKindAPI
	case "connector":
		return ResourceKindConnector
	default:
		return kind
	}
}

// ResourceKindFromCompiler converts a compiler resource kind to a runtime resource kind.
func ResourceKindFromCompiler(kind compilerv1.ResourceKind) string {
	switch kind {
	case compilerv1.ResourceKindSource:
		return ResourceKindSource
	case compilerv1.ResourceKindModel:
		return ResourceKindModel
	case compilerv1.ResourceKindMetricsView:
		return ResourceKindMetricsView
	case compilerv1.ResourceKindExplore:
		return ResourceKindExplore
	case compilerv1.ResourceKindMigration:
		return ResourceKindMigration
	case compilerv1.ResourceKindReport:
		return ResourceKindReport
	case compilerv1.ResourceKindAlert:
		return ResourceKindAlert
	case compilerv1.ResourceKindTheme:
		return ResourceKindTheme
	case compilerv1.ResourceKindComponent:
		return ResourceKindComponent
	case compilerv1.ResourceKindCanvas:
		return ResourceKindCanvas
	case compilerv1.ResourceKindAPI:
		return ResourceKindAPI
	case compilerv1.ResourceKindConnector:
		return ResourceKindConnector
	default:
		panic(fmt.Errorf("unknown compiler resource type %q", kind))
	}
}

// ResourceKindToCompiler converts a runtime resource kind to a compiler resource kind.
func ResourceKindToCompiler(kind string) compilerv1.ResourceKind {
	switch kind {
	case ResourceKindSource:
		return compilerv1.ResourceKindSource
	case ResourceKindModel:
		return compilerv1.ResourceKindModel
	case ResourceKindMetricsView:
		return compilerv1.ResourceKindMetricsView
	case ResourceKindExplore:
		return compilerv1.ResourceKindExplore
	case ResourceKindMigration:
		return compilerv1.ResourceKindMigration
	case ResourceKindReport:
		return compilerv1.ResourceKindReport
	case ResourceKindAlert:
		return compilerv1.ResourceKindAlert
	case ResourceKindTheme:
		return compilerv1.ResourceKindTheme
	case ResourceKindComponent:
		return compilerv1.ResourceKindComponent
	case ResourceKindCanvas:
		return compilerv1.ResourceKindCanvas
	case ResourceKindAPI:
		return compilerv1.ResourceKindAPI
	case ResourceKindConnector:
		return compilerv1.ResourceKindConnector
	case ResourceKindProjectParser, ResourceKindPullTrigger, ResourceKindRefreshTrigger, ResourceKindBucketPlanner:
		panic(fmt.Errorf("unsupported resource type %q", kind))
	default:
		panic(fmt.Errorf("unknown resource type %q", kind))
	}
}

// ResourceNameFromCompiler converts a compiler resource name to a runtime resource name.
func ResourceNameFromCompiler(name compilerv1.ResourceName) *runtimev1.ResourceName {
	return &runtimev1.ResourceName{Kind: ResourceKindFromCompiler(name.Kind), Name: name.Name}
}

// ResourceNameToCompiler converts a runtime resource name to a compiler resource name.
func ResourceNameToCompiler(name *runtimev1.ResourceName) compilerv1.ResourceName {
	return compilerv1.ResourceName{Kind: ResourceKindToCompiler(name.Kind), Name: name.Name}
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
