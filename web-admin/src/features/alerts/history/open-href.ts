import {
  V1AssertionStatus,
  type V1AlertExecution,
} from "@rilldata/web-common/runtime-client/gen/index.schemas";

/**
 * Where a check in the alert's history opens: the data it saw, through the same open page
 * as the alert email's link. Only checks that triggered the alert have data to show.
 */
export function alertExecutionOpenHref(
  organization: string,
  project: string,
  alert: string,
  execution: V1AlertExecution,
): string | undefined {
  if (
    execution.result?.status !== V1AssertionStatus.ASSERTION_STATUS_FAIL ||
    !execution.executionTime
  ) {
    return undefined;
  }
  const params = new URLSearchParams({
    execution_time: execution.executionTime,
  });
  return `/${organization}/${project}/-/alerts/${encodeURIComponent(alert)}/open?${params}`;
}
