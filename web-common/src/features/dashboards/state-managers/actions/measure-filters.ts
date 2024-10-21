import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import type { DashboardMutables } from "@rilldata/web-common/features/dashboards/state-managers/actions/types";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";

export function setMeasureFilter(
  { dashboard }: DashboardMutables,
  dimensionName: string,
  filter: MeasureFilterEntry,
) {
  if (dashboard.temporaryFilterName !== null) {
    dashboard.temporaryFilterName = null;
  }

  const dimId = dashboard.dimensionThresholdFilters.findIndex(
    (dtf) => dtf.name === dimensionName,
  );
  let dimThresholdFilter: DimensionThresholdFilter;
  if (dimId === -1) {
    dimThresholdFilter = {
      name: dimensionName,
      filters: [],
    };
    dashboard.dimensionThresholdFilters.push(dimThresholdFilter);
  } else {
    dimThresholdFilter = dashboard.dimensionThresholdFilters[dimId];
  }

  const exprIdx = dimThresholdFilter.filters.findIndex(
    (f) => f.measure === filter.measure,
  );
  if (exprIdx === -1) {
    // if there is no expression for the measure push to the end
    dimThresholdFilter.filters.push(filter);
  } else if (exprIdx >= 0) {
    // else replace the existing measure filter
    dimThresholdFilter.filters.splice(exprIdx, 1, filter);
  }
}

export function removeMeasureFilter(
  { dashboard }: DashboardMutables,
  dimensionName: string,
  measureName: string,
) {
  if (dashboard.temporaryFilterName === measureName) {
    dashboard.temporaryFilterName = null;
    return;
  }

  const dimId = dashboard.dimensionThresholdFilters.findIndex(
    (dtf) => dtf.name === dimensionName,
  );
  if (dimId === -1) return;
  const dimThresholdFilter = dashboard.dimensionThresholdFilters[dimId];

  const exprIdx = dimThresholdFilter.filters.findIndex(
    (f) => f.measure === measureName,
  );
  if (exprIdx === -1) return;
  dimThresholdFilter.filters.splice(exprIdx, 1);

  // if dimension threshold filter is empty remove it
  if (dimThresholdFilter.filters.length === 0) {
    dashboard.dimensionThresholdFilters.splice(dimId, 1);
  }
}

export const measureFilterActions = {
  setMeasureFilter,

  removeMeasureFilter,
};
