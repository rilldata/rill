import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { isResourceReconciling } from "../../../../lib/refetch-interval-store";

export type ResourceStatus = "ok" | "pending" | "warning" | "errored";

const TEST_FAILURE_MARKER = "tests failed:";

/**
 * Determines the display status of a resource based on its reconcile state.
 * - "errored": has a reconcile error (excluding test-only failures)
 * - "warning": has test-only failures
 * - "pending": reconcile is PENDING or RUNNING
 * - "ok": otherwise (IDLE, UNSPECIFIED, etc.)
 */
export function getResourceStatus(r: V1Resource): ResourceStatus {
  const error = r.meta?.reconcileError ?? "";
  if (error && !error.includes(TEST_FAILURE_MARKER)) return "errored";
  if (error && error.includes(TEST_FAILURE_MARKER)) return "warning";
  if (isResourceReconciling(r)) return "pending";
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
