import type { DashboardMutables } from "@rilldata/web-common/features/dashboards/state-managers/actions/types";
import { getWhereFilterExpressionIndex } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

export function toggleMeasureFilter(
  { dashboard, cancelQueries }: DashboardMutables,
  measureName: string,
  filter?: V1Expression
) {
  // if we are able to update the filters, we must cancel any queries
  // that are currently running.
  cancelQueries();

  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(measureName);

  if (exprIdx === undefined || exprIdx === -1) {
    if (filter !== undefined) {
      dashboard.havingFilter.cond?.exprs?.push(filter);
    }
  } else if (exprIdx >= 0) {
    dashboard.havingFilter.cond?.exprs?.splice(exprIdx, 1);
  }
}

export const measureFilterActions = {
  toggleMeasureFilter,
};
