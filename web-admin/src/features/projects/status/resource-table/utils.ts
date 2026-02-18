import {
  V1ReconcileStatus,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";

export type ResourceStatus = "error" | "warn" | "ok";

/**
 * Determines the display status of a resource based on its reconcile state.
 * - "error": has a reconcile error
 * - "warn": reconcile is PENDING or RUNNING
 * - "ok": otherwise (IDLE, UNSPECIFIED, etc.)
 */
export function getResourceStatus(r: V1Resource): ResourceStatus {
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
 * Filters resources by kind, search text, and status.
 * All filters are AND-ed together. Empty filter arrays match all.
 */
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
