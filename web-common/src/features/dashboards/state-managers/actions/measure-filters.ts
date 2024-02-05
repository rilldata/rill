import type { DashboardMutables } from "@rilldata/web-common/features/dashboards/state-managers/actions/types";
import {
  getHavingFilterExpressionIndex,
  getMeasureFilterForDimensionIndex,
} from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

export function setMeasureFilter(
  { dashboard, cancelQueries }: DashboardMutables,
  dimensionName: string,
  measureName: string,
  filter: V1Expression,
) {
  // if we are able to update the filters, we must cancel any queries
  // that are currently running.
  cancelQueries();

  if (dashboard.temporaryFilterName !== null) {
    dashboard.temporaryFilterName = null;
  }

  const dimId = getMeasureFilterForDimensionIndex({ dashboard })(dimensionName);
  let dimThresholdFilter: DimensionThresholdFilter;
  if (dimId === -1) {
    dimThresholdFilter = {
      name: dimensionName,
      filter: createAndExpression([]),
    };
    dashboard.dimensionThresholdFilters.push(dimThresholdFilter);
  } else {
    dimThresholdFilter = dashboard.dimensionThresholdFilters[dimId];
  }

  const exprIdx = getHavingFilterExpressionIndex(
    dimThresholdFilter.filter,
    measureName,
  );
  if (exprIdx === -1) {
    // if there is no expression for the measure push to the end
    dimThresholdFilter.filter.cond?.exprs?.push(filter);
  } else if (exprIdx >= 0) {
    // else replace the existing measure filter
    dimThresholdFilter.filter.cond?.exprs?.splice(exprIdx, 1, filter);
  }
}

export function removeMeasureFilter(
  { dashboard, cancelQueries }: DashboardMutables,
  dimensionName: string,
  measureName: string,
) {
  // if we are able to update the filters, we must cancel any queries
  // that are currently running.
  cancelQueries();

  if (dashboard.temporaryFilterName === measureName) {
    dashboard.temporaryFilterName = null;
    return;
  }

  const dimId = getMeasureFilterForDimensionIndex({ dashboard })(dimensionName);
  if (dimId === -1) return;
  const dimThresholdFilter = dashboard.dimensionThresholdFilters[dimId];

  const exprIdx = getHavingFilterExpressionIndex(
    dimThresholdFilter.filter,
    measureName,
  );
  if (exprIdx === -1) return;
  dimThresholdFilter.filter.cond?.exprs?.splice(exprIdx, 1);

  // if dimension threshold filter is empty remove it
  if (dimThresholdFilter.filter.cond?.exprs?.length === 0) {
    dashboard.dimensionThresholdFilters.splice(dimId, 1);
  }
}

export const measureFilterActions = {
  setMeasureFilter,

  removeMeasureFilter,
};
