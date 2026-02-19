import {
  V1ReconcileStatus,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

export type StatusFilter = { label: string; value: string };

export const statusFilters: StatusFilter[] = [
  { label: "Error", value: "error" },
  { label: "Warn", value: "warn" },
  { label: "OK", value: "ok" },
];

export const filterableTypes = [
  ResourceKind.Source,
  ResourceKind.Model,
  ResourceKind.MetricsView,
  ResourceKind.Explore,
  ResourceKind.Canvas,
  ResourceKind.Theme,
  ResourceKind.Report,
  ResourceKind.Alert,
  ResourceKind.API,
  ResourceKind.Connector,
];

export function getResourceStatus(r: V1Resource): string {
  if (r.meta?.reconcileError) return "error";
  const status = r.meta?.reconcileStatus;
  if (
    status === V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
    status === V1ReconcileStatus.RECONCILE_STATUS_RUNNING
  )
    return "warn";
  return "ok";
}

/**
 * Returns a numeric priority for a reconcile status, used for table sorting.
 * Higher values sort first: Running (4) > Pending (3) > Idle (2) > Unknown (1)
 */
export function getStatusPriority(
  status: V1ReconcileStatus | undefined,
): number {
  switch (status) {
    case V1ReconcileStatus.RECONCILE_STATUS_RUNNING:
      return 4;
    case V1ReconcileStatus.RECONCILE_STATUS_PENDING:
      return 3;
    case V1ReconcileStatus.RECONCILE_STATUS_IDLE:
      return 2;
    case V1ReconcileStatus.RECONCILE_STATUS_UNSPECIFIED:
    default:
      return 1;
  }
}

export function filterResources(
  resources: V1Resource[] | undefined,
  types: string[],
  search: string,
  statuses: string[],
): V1Resource[] {
  if (!resources) return [];

  return resources.filter((r) => {
    const kind = r.meta?.name?.kind;
    const name = r.meta?.name?.name ?? "";

    const matchesType = types.length === 0 || types.includes(kind ?? "");
    const matchesSearch =
      !search || name.toLowerCase().includes(search.toLowerCase());
    const matchesStatus =
      statuses.length === 0 || statuses.includes(getResourceStatus(r));

    return matchesType && matchesSearch && matchesStatus;
  });
}
