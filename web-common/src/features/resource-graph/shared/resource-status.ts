import type { V1Resource } from "@rilldata/web-common/runtime-client";
import type { ResourceStatusFilterValue } from "./types";

export const TEST_FAILURE_MARKER = "tests failed:";

/**
 * Determines the display status of a resource based on its reconcile state.
 * - "errored": has a reconcile error (excluding test-only failures)
 * - "warning": has test-only failures
 * - "pending": reconcile is PENDING or RUNNING
 * - "ok": otherwise (IDLE, UNSPECIFIED, etc.)
 */
export function getResourceStatus(r: V1Resource): ResourceStatusFilterValue {
  const error = r.meta?.reconcileError ?? "";
  if (error && !error.includes(TEST_FAILURE_MARKER)) return "errored";
  if (error && error.includes(TEST_FAILURE_MARKER)) return "warning";
  if (
    r.meta?.reconcileStatus === "RECONCILE_STATUS_PENDING" ||
    r.meta?.reconcileStatus === "RECONCILE_STATUS_RUNNING"
  )
    return "pending";
  return "ok";
}
