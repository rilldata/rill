import type {
  MetricsViewSpecDimensionV2,
  RpcStatus,
  V1MetricsViewTotalsResponse,
} from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import { prepareVirtualizedDimTableColumns } from "../../dimension-table/dimension-table-utils";
import type { VirtualizedTableColumns } from "@rilldata/web-local/lib/types";
import { visibleMeasures } from "./measures";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import { isSummableMeasure } from "../../dashboard-utils";
import { isTimeComparisonActive, timeControlsState } from "./time-range";
import { isValidPercentOfTotal } from "./active-measure";

export const primaryDimension = (
  dashData: DashboardDataSources
): MetricsViewSpecDimensionV2 | undefined => {
  const dimName = dashData.dashboard.selectedDimensionName;
  return dashData.metricsSpecQueryResult.data?.dimensions?.find(
    (dim) => dim.name === dimName
  );
};

export const virtualizedTableColumns =
  (
    dashData: DashboardDataSources
  ): ((
    totalsQuery: QueryObserverResult<V1MetricsViewTotalsResponse, RpcStatus>
  ) => VirtualizedTableColumns[]) =>
  (totalsQuery) => {
    const dimension = primaryDimension(dashData);

    timeControlsState(dashData).showComparison;
    if (!dimension) return [];

    const measures = visibleMeasures(dashData);

    const referenceValues: { [key: string]: number } = {};
    if (totalsQuery?.data?.data) {
      measures.map((m) => {
        if (m.name && isSummableMeasure(m)) {
          referenceValues[m.name] = totalsQuery.data?.data?.[m.name];
        }
      });
    }

    return prepareVirtualizedDimTableColumns(
      dashData.dashboard,
      measures,
      referenceValues,
      dimension,
      isTimeComparisonActive(dashData),
      isValidPercentOfTotal(dashData)
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
};
