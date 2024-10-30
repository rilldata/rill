import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { getIndependentMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  QueryServiceMetricsViewAggregationBody,
  V1Expression,
} from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import { prepareSortedQueryBody } from "../../dashboard-utils";
import { activeMeasureName, selectedMeasureNames } from "./active-measure";
import { sortingSelectors } from "./sorting";
import { timeControlsState } from "./time-range";
import {
  getFiltersForOtherDimensions,
  additionalMeasures,
} from "../../selectors";
import { updateFilterOnSearch } from "../../dimension-table/dimension-table-utils";
import { dimensionTableSearchString } from "./dimension-table";

/**
 * Returns the sorted query body for the dimension table for the
 * active dimension.
 *
 * Safety: this readable should ONLY be used when the active dimension
 * is not undefined, ie then the dimension table is visible.
 */
export function dimensionTableSortedQueryBody(
  dashData: DashboardDataSources,
): QueryServiceMetricsViewAggregationBody {
  const dimensionName = dashData.dashboard.selectedDimensionName;
  if (!dimensionName) {
    return {};
  }
  let filters: V1Expression | undefined = getFiltersForOtherDimensions(
    dashData.dashboard.whereFilter,
    dimensionName,
  );
  const searchString = dimensionTableSearchString(dashData);
  if (searchString !== undefined) {
    filters = updateFilterOnSearch(filters, searchString, dimensionName);
  }

  return prepareSortedQueryBody(
    dimensionName,
    measuresForDimensionTable(dashData),
    timeControlsState(dashData),
    sortingSelectors.sortMeasure(dashData),
    sortingSelectors.sortType(dashData),
    sortingSelectors.sortedAscending(dashData),
    mergeMeasureFilters(dashData.dashboard, filters),
    250,
  );
}

function measuresForDimensionTable(dashData: DashboardDataSources) {
  const allMeasures = new Set([
    ...selectedMeasureNames(dashData),
    ...additionalMeasures(
      dashData.dashboard.leaderboardMeasureName,
      dashData.dashboard.dimensionThresholdFilters,
    ),
  ]);
  return getIndependentMeasures(dashData.validMetricsView ?? {}, [
    ...allMeasures,
  ]);
}

export function dimensionTableTotalQueryBody(
  dashData: DashboardDataSources,
): QueryServiceMetricsViewAggregationBody {
  const dimensionName = dashData.dashboard.selectedDimensionName;
  if (!dimensionName) {
    return {};
  }

  return {
    measures: [{ name: activeMeasureName(dashData) }],
    where: sanitiseExpression(
      mergeMeasureFilters(
        dashData.dashboard,
        getFiltersForOtherDimensions(
          dashData.dashboard.whereFilter,
          dimensionName,
        ),
      ),
      undefined,
    ),
    timeStart: timeControlsState(dashData).timeStart,
    timeEnd: timeControlsState(dashData).timeEnd,
  };
}

export const leaderboardQuerySelectors = {
  /**
   * Readable containing the sorted query body for the dimension table.
   */
  dimensionTableSortedQueryBody,

  /**
   * Readable containing the totals query body for the dimension table.
   */
  dimensionTableTotalQueryBody,
};
