import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  type TimeControlState,
  useTimeControlStore,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { mapSelectedTimeRangeToV1TimeRange } from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import type {
  V1ExploreSpec,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationRequest,
  V1MetricsViewSpec,
  V1Query,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";
import { buildWhereParamForDimensionTableAndTDDExports } from "../../exports/export-filters";
import { dimensionSearchText as dimensionSearchTextStore } from "../stores/dashboard-stores";

export function getTDDExportQuery(
  ctx: StateManagers,
  isScheduled: boolean,
): V1Query {
  const metricsViewName = get(ctx.metricsViewName);
  const dashboardState = get(ctx.dashboardStore);
  const timeControlState = get(useTimeControlStore(ctx));
  const validSpec = get(ctx.validSpecStore);
  const dimensionSearchText = get(dimensionSearchTextStore);

  const query: V1Query = {
    metricsViewAggregationRequest: getTDDAggregationRequest(
      metricsViewName,
      dashboardState,
      timeControlState,
      validSpec.data?.metricsView,
      validSpec.data?.explore,
      dimensionSearchText,
      isScheduled,
    ),
  };

  return query;
}

function getTDDAggregationRequest(
  metricsViewName: string,
  exploreState: ExploreState,
  timeControlState: TimeControlState,
  metricsView: V1MetricsViewSpec | undefined,
  explore: V1ExploreSpec | undefined,
  dimensionSearchText: string,
  isScheduled: boolean,
): undefined | V1MetricsViewAggregationRequest {
  if (
    !metricsView ||
    !explore ||
    !timeControlState.ready ||
    !exploreState.tdd.expandedMeasureName
  )
    return undefined;

  let timeRange: V1TimeRange | undefined;
  if (isScheduled) {
    timeRange = mapSelectedTimeRangeToV1TimeRange(
      timeControlState.selectedTimeRange,
      exploreState.selectedTimezone,
      explore,
    );
  } else {
    timeRange = {
      start: timeControlState.timeStart,
      end: timeControlState.timeEnd,
    };
  }
  if (!timeRange) return undefined;

  const measures: V1MetricsViewAggregationMeasure[] = [
    { name: exploreState.tdd.expandedMeasureName },
  ];

  // CAST SAFETY: exports are only available in TDD when a comparison dimension is selected
  const dimensionName = exploreState.selectedComparisonDimension as string;
  const timeDimension = metricsView.timeDimension ?? "";

  return {
    instanceId: get(runtime).instanceId,
    metricsView: metricsViewName,
    dimensions: [
      { name: dimensionName },
      {
        name: metricsView.timeDimension ?? "",
        timeGrain: exploreState.selectedTimeRange?.interval,
        timeZone: exploreState.selectedTimezone,
      },
    ],
    measures,
    timeRange,
    pivotOn: [timeDimension],
    sort: [
      {
        name: dimensionName,
        desc: exploreState.sortDirection === SortDirection.DESCENDING,
      },
    ],
    where: buildWhereParamForDimensionTableAndTDDExports(
      exploreState.whereFilter,
      exploreState.dimensionThresholdFilters,
      dimensionName,
      dimensionSearchText,
    ),
    offset: "0",
  };
}
