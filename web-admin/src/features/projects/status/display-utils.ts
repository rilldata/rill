import type { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";

export function prettyReconcileStatus(status: V1ReconcileStatus) {
  switch (status) {
    case "RECONCILE_STATUS_IDLE":
      return "Idle";
    case "RECONCILE_STATUS_PENDING":
      return "Pending";
    case "RECONCILE_STATUS_RUNNING":
      return "Running";
    case "RECONCILE_STATUS_UNSPECIFIED":
      return "Unspecified";
  }
}

export function getResourceKindTagColor(kind: string) {
  switch (kind) {
    case "rill.runtime.v1.MetricsView":
      return "blue";
    case "rill.runtime.v1.Model":
      return "green";
    case "rill.runtime.v1.Report":
      return "purple";
    case "rill.runtime.v1.Source":
      return "orange";
    case "rill.runtime.v1.Theme":
      return "yellow";
    default:
      return "gray";
  }
}
