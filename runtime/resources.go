package runtime

import (
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/parser"
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
// If the kind doesn't match a known shorthand, it is returned as-is.
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

// ResourceKindFromParser converts a parser resource kind to a runtime resource kind.
func ResourceKindFromParser(kind parser.ResourceKind) string {
	switch kind {
	case parser.ResourceKindSource:
		return ResourceKindSource
	case parser.ResourceKindModel:
		return ResourceKindModel
	case parser.ResourceKindMetricsView:
		return ResourceKindMetricsView
	case parser.ResourceKindExplore:
		return ResourceKindExplore
	case parser.ResourceKindMigration:
		return ResourceKindMigration
	case parser.ResourceKindReport:
		return ResourceKindReport
	case parser.ResourceKindAlert:
		return ResourceKindAlert
	case parser.ResourceKindTheme:
		return ResourceKindTheme
	case parser.ResourceKindComponent:
		return ResourceKindComponent
	case parser.ResourceKindCanvas:
		return ResourceKindCanvas
	case parser.ResourceKindAPI:
		return ResourceKindAPI
	case parser.ResourceKindConnector:
		return ResourceKindConnector
	default:
		panic(fmt.Errorf("unknown parser resource type %q", kind))
	}
}

// ResourceKindToParser converts a runtime resource kind to a parser resource kind.
func ResourceKindToParser(kind string) parser.ResourceKind {
	switch kind {
	case ResourceKindSource:
		return parser.ResourceKindSource
	case ResourceKindModel:
		return parser.ResourceKindModel
	case ResourceKindMetricsView:
		return parser.ResourceKindMetricsView
	case ResourceKindExplore:
		return parser.ResourceKindExplore
	case ResourceKindMigration:
		return parser.ResourceKindMigration
	case ResourceKindReport:
		return parser.ResourceKindReport
	case ResourceKindAlert:
		return parser.ResourceKindAlert
	case ResourceKindTheme:
		return parser.ResourceKindTheme
	case ResourceKindComponent:
		return parser.ResourceKindComponent
	case ResourceKindCanvas:
		return parser.ResourceKindCanvas
	case ResourceKindAPI:
		return parser.ResourceKindAPI
	case ResourceKindConnector:
		return parser.ResourceKindConnector
	case ResourceKindProjectParser, ResourceKindPullTrigger, ResourceKindRefreshTrigger, ResourceKindBucketPlanner:
		panic(fmt.Errorf("unsupported resource type %q", kind))
	default:
		panic(fmt.Errorf("unknown resource type %q", kind))
	}
}

// ResourceNameFromParser converts a parser resource name to a runtime resource name.
func ResourceNameFromParser(name parser.ResourceName) *runtimev1.ResourceName {
	return &runtimev1.ResourceName{Kind: ResourceKindFromParser(name.Kind), Name: name.Name}
}

// ResourceNameToParser converts a runtime resource name to a parser resource name.
func ResourceNameToParser(name *runtimev1.ResourceName) parser.ResourceName {
	return parser.ResourceName{Kind: ResourceKindToParser(name.Kind), Name: name.Name}
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
