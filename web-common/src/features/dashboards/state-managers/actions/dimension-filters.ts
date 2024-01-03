import {
  createInExpression,
  getValueIndexInExpression,
  negateExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-generators";
import type { DashboardMutables } from "./types";
import { getWhereFilterExpressionIndex } from "../selectors/dimension-filters";

export function toggleDimensionValueSelection(
  { dashboard, cancelQueries }: DashboardMutables,
  dimensionName: string,
  dimensionValue: string
) {
  // if we are able to update the filters, we must cancel any queries
  // that are currently running.
  cancelQueries();

  console.log(dashboard.whereFilter);
  const isInclude = !dashboard.dimensionFilterExcludeMode.get(dimensionName);
  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(dimensionName);
  if (exprIdx === undefined || exprIdx === -1) {
    dashboard.whereFilter.cond?.exprs?.push(
      createInExpression(dimensionName, [dimensionValue], !isInclude)
    );
    return;
  }

  const expr = dashboard.whereFilter.cond?.exprs?.[exprIdx];
  if (!expr?.cond?.exprs) {
    // should never happen since getWhereFilterExpressionIndex runs a find
    return;
  }

  const inIdx = getValueIndexInExpression(expr, dimensionValue) as number;
  if (inIdx === -1) {
    expr.cond.exprs.push({ val: dimensionValue });
  } else {
    expr.cond.exprs.splice(inIdx, 1);
    // Only decrement pinIndex if the removed value was before the pinned value
    if (dashboard.pinIndex >= inIdx) {
      dashboard.pinIndex--;
    }
    // remove the dimension entry if all values are removed
    if (expr.cond.exprs.length === 1) {
      dashboard.whereFilter.cond?.exprs?.splice(exprIdx, 1);
    }
  }
}

export function toggleDimensionNameSelection(
  { dashboard, cancelQueries }: DashboardMutables,
  dimensionName: string
) {
  // if we are able to update the filters, we must cancel any queries
  // that are currently running.
  cancelQueries();

  const isExclude =
    dashboard.dimensionFilterExcludeMode.get(dimensionName) ?? false;
  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(dimensionName);
  if (exprIdx === undefined || exprIdx === -1) {
    // if filter for dimension exist add it
    dashboard.whereFilter.cond?.exprs?.push(
      createInExpression(dimensionName, [], isExclude)
    );
  } else {
    // else remove it
    dashboard.whereFilter?.cond?.exprs?.splice(exprIdx, 1);
  }
}

export function toggleDimensionFilterMode(
  { dashboard, cancelQueries }: DashboardMutables,
  dimensionName: string
) {
  const exclude = dashboard.dimensionFilterExcludeMode.get(dimensionName);
  dashboard.dimensionFilterExcludeMode.set(dimensionName, !exclude);

  if (!dashboard.whereFilter?.cond?.exprs) {
    return;
  }

  // if we are able to update the filters, we must cancel any queries
  // that are currently running.
  cancelQueries();

  const exprIdx = dashboard.whereFilter.cond.exprs.findIndex(
    (e) => e.cond?.exprs?.[0].ident === dimensionName
  );
  if (exprIdx === -1) {
    return;
  }
  dashboard.whereFilter.cond.exprs[exprIdx] = negateExpression(
    dashboard.whereFilter.cond.exprs[exprIdx]
  );
}

export const dimensionFilterActions = {
  /**
   * Toggles whether the given dimension value is selected in the
   * dimension filter for the given dimension.
   *
   * Note that this is different than the include/exclude mode for
   * dimension filters. This is a toggle for a specific value, whereas
   * the include/exclude mode is a toggle for the entire dimension.
   */
  toggleDimensionValueSelection,
  toggleDimensionNameSelection,
  toggleDimensionFilterMode,
};
