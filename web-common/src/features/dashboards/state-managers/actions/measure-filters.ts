import type { DashboardMutables } from "@rilldata/web-common/features/dashboards/state-managers/actions/types";
import { getHavingFilterExpressionIndex } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

export function toggleMeasureFilter(
  { dashboard, cancelQueries }: DashboardMutables,
  measureName: string,
  filter?: V1Expression
) {
  // if we are able to update the filters, we must cancel any queries
  // that are currently running.
  cancelQueries();

  const exprIdx = getHavingFilterExpressionIndex({ dashboard })(measureName);

  if (exprIdx === undefined || exprIdx === -1) {
    if (filter !== undefined) {
      dashboard.havingFilter.cond?.exprs?.push(filter);
    }
  } else if (exprIdx >= 0) {
    dashboard.havingFilter.cond?.exprs?.splice(exprIdx, 1);
  }
}

export function setMeasureFilter(
  { dashboard, cancelQueries }: DashboardMutables,
  measureName: string,
  filter: V1Expression
) {
  // if we are able to update the filters, we must cancel any queries
  // that are currently running.
  cancelQueries();

  const exprIdx = getHavingFilterExpressionIndex({ dashboard })(measureName);
  console.log(dashboard.havingFilter, exprIdx);
  if (exprIdx === undefined || exprIdx === -1) {
    dashboard.havingFilter.cond?.exprs?.push(filter);
  } else if (exprIdx >= 0) {
    dashboard.havingFilter.cond?.exprs?.splice(exprIdx, 1, filter);
  }
}

export const measureFilterActions = {
  toggleMeasureFilter,
  setMeasureFilter,
};
