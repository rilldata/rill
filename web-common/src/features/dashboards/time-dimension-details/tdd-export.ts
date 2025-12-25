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
import httpClient from "@rilldata/web-common/runtime-client/http-client";
import { get } from "svelte/store";
import { buildWhereParamForDimensionTableAndTDDExports } from "../../exports/export-filters";
import { dimensionSearchText as dimensionSearchTextStore } from "../stores/dashboard-stores";

export function getTDDExportQuery(
  ctx: StateManagers,
  isScheduled: boolean,
): V1Query {
  const metricsViewName = get(ctx.metricsViewName);
  const exploreState = get(ctx.dashboardStore);
  const timeControlState = get(useTimeControlStore(ctx));
  const validSpec = get(ctx.validSpecStore);
  const dimensionSearchText = get(dimensionSearchTextStore);

  const query: V1Query = {
    metricsViewAggregationRequest: getTDDAggregationRequest({
      metricsViewName,
      exploreState,
      timeControlState,
      metricsViewSpec: validSpec.data?.metricsView,
      exploreSpec: validSpec.data?.explore,
      dimensionSearchText,
      isScheduled,
    }),
  };

  return query;
}

export function getTDDAggregationRequest({
  metricsViewName,
  exploreState,
  timeControlState,
  metricsViewSpec,
  exploreSpec,
  dimensionSearchText,
  isScheduled,
}: {
  metricsViewName: string;
  exploreState: ExploreState;
  timeControlState: TimeControlState;
  metricsViewSpec: V1MetricsViewSpec | undefined;
  exploreSpec: V1ExploreSpec | undefined;
  dimensionSearchText: string;
  isScheduled: boolean;
}): undefined | V1MetricsViewAggregationRequest {
  if (
    !metricsViewSpec ||
    !exploreSpec ||
    !timeControlState.ready ||
    !exploreState.tdd.expandedMeasureName
  )
    return undefined;

  let timeRange: V1TimeRange | undefined;
  if (isScheduled) {
    timeRange = mapSelectedTimeRangeToV1TimeRange(
      timeControlState.selectedTimeRange,
      exploreState.selectedTimezone,
      exploreSpec,
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
  const timeDimension = metricsViewSpec.timeDimension ?? "";

  return {
    instanceId: httpClient.getInstanceId(),
    metricsView: metricsViewName,
    dimensions: [
      { name: dimensionName },
      {
        name: metricsViewSpec.timeDimension ?? "",
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
