import type {
  MetricsViewSpecDimensionV2,
  RpcStatus,
  V1MetricsViewComparisonResponse,
  V1MetricsViewTotalsResponse,
} from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import {
  prepareDimensionTableRows,
  prepareVirtualizedDimTableColumns,
} from "../../dimension-table/dimension-table-utils";
import type { VirtualizedTableColumns } from "@rilldata/web-local/lib/types";
import { allMeasures, visibleMeasures } from "./measures";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import { getDimensionColumn, isSummableMeasure } from "../../dashboard-utils";
import { isTimeComparisonActive } from "./time-range";
import { activeMeasureName, isValidPercentOfTotal } from "./active-measure";
import { selectedDimensionValues } from "./dimension-filters";
import type { DimensionTableRow } from "../../dimension-table/dimension-table-types";

export const selectedDimensionValueNames = (
  dashData: DashboardDataSources,
): string[] => {
  const dimension = dashData.dashboard.selectedDimensionName;
  if (!dimension) return [];
  return selectedDimensionValues(dashData)(dimension);
};

export const primaryDimension = (
  dashData: DashboardDataSources,
): MetricsViewSpecDimensionV2 | undefined => {
  const dimName = dashData.dashboard.selectedDimensionName;
  return dashData.metricsSpecQueryResult.data?.dimensions?.find(
    (dim) => dim.name === dimName,
  );
};

export const dimensionTableSearchString = (
  dashData: DashboardDataSources,
): string | undefined => dashData.dashboard.dimensionSearchText;

export const virtualizedTableColumns =
  (
    dashData: DashboardDataSources,
  ): ((
    totalsQuery: QueryObserverResult<V1MetricsViewTotalsResponse, RpcStatus>,
  ) => VirtualizedTableColumns[]) =>
  (totalsQuery) => {
    const dimension = primaryDimension(dashData);

    if (!dimension) return [];

    const measures = visibleMeasures(dashData);

    const measureTotals: { [key: string]: number } = {};
    if (totalsQuery?.data?.data) {
      measures.map((m) => {
        if (m.name && isSummableMeasure(m)) {
          measureTotals[m.name] = totalsQuery.data?.data?.[m.name];
        }
      });
    }

    return prepareVirtualizedDimTableColumns(
      dashData.dashboard,
      measures,
      measureTotals,
      dimension,
      isTimeComparisonActive(dashData),
      isValidPercentOfTotal(dashData),
    );
  };

export const prepareDimTableRows =
  (
    dashData: DashboardDataSources,
  ): ((
    sortedQuery: QueryObserverResult<
      V1MetricsViewComparisonResponse,
      RpcStatus
    >,
    unfilteredTotal: number,
  ) => DimensionTableRow[]) =>
  (sortedQuery, unfilteredTotal) => {
    const dimension = primaryDimension(dashData);

    if (!dimension) return [];

    const dimensionColumn = getDimensionColumn(dimension);
    const leaderboardMeasureName = activeMeasureName(dashData);

    // FIXME: should this really be all measures, or just visible measures?
    const measures = allMeasures(dashData);

    return prepareDimensionTableRows(
      sortedQuery?.data?.rows ?? [],
      measures,
      leaderboardMeasureName,
      dimensionColumn,
      isTimeComparisonActive(dashData),
      isValidPercentOfTotal(dashData),
      unfilteredTotal,
    );
  };

export const dimensionTableSelectors = {
  /**
   * gets the VirtualizedTableColumns array for the dimension table.
   */
  virtualizedTableColumns,

  /**
   * gets the MetricsViewSpecDimensionV2 for the dimension table's
   * primary dimension.
   */
  primaryDimension,

  /**
   * gets the names of the selected dimension values for the primary dimension.
   */
  selectedDimensionValueNames,

  /**
   * A readable containaing a function that will prepare
   * the dimension table rows for given a sorted query
   * and unfiltered total.
   */
  prepareDimTableRows,

  /**
   * gets the dimension table search string.
   */
  dimensionTableSearchString,
};
