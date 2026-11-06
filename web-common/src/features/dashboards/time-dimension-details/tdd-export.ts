import {
  applyExactMeasures,
  mapEphemeralMeasuresForRequest,
} from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
import { requestExactNames } from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
import { exploreExactSurface } from "@rilldata/web-common/features/dashboards/exact-measures/explore-surfaces";
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

  // A one-shot export downloads what is on screen, so it carries the viewer's
  // exact selection. A scheduled report must not: exact values are only
  // fetched after an explicit click, never on a schedule.
  const exactMeasures =
    isScheduled || !exploreState.tdd.expandedMeasureName
      ? undefined
      : requestExactNames(
          [exploreState.tdd.expandedMeasureName],
          get(ctx.selectors.exactMeasures.exactMeasureContext)(
            exploreExactSurface.tdd,
          ),
        );

  const query: V1Query = {
    metricsViewAggregationRequest: getTDDAggregationRequest({
      instanceId: ctx.runtimeClient.instanceId,
      metricsViewName,
      exploreState,
      timeControlState,
      metricsViewSpec: validSpec.data?.metricsView,
      exploreSpec: validSpec.data?.explore,
      dimensionSearchText,
      isScheduled,
      exactMeasures,
    }),
  };

  return query;
}

export function getTDDAggregationRequest({
  instanceId,
  metricsViewName,
  exploreState,
  timeControlState,
  metricsViewSpec,
  exploreSpec,
  dimensionSearchText,
  isScheduled,
  exactMeasures,
}: {
  instanceId: string;
  metricsViewName: string;
  exploreState: ExploreState;
  timeControlState: TimeControlState;
  metricsViewSpec: V1MetricsViewSpec | undefined;
  exploreSpec: V1ExploreSpec | undefined;
  dimensionSearchText: string;
  isScheduled: boolean;
  // Measures to request with `exact: true`; see `requestExactNames`.
  exactMeasures?: ReadonlySet<string>;
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

  const measures: V1MetricsViewAggregationMeasure[] = applyExactMeasures(
    mapEphemeralMeasuresForRequest(
      [{ name: exploreState.tdd.expandedMeasureName }],
      exploreState.ephemeralMeasures,
    ),
    exactMeasures,
  );

  // CAST SAFETY: exports are only available in TDD when a comparison dimension is selected
  const dimensionName = exploreState.selectedComparisonDimension as string;
  const timeDimension =
    exploreState.selectedTimeDimension || metricsViewSpec.timeDimension || "";

  return {
    instanceId,
    metricsView: metricsViewName,
    dimensions: [
      { name: dimensionName },
      {
        name: timeDimension,
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
      dimensionName,
      dimensionSearchText,
    ),
    offset: "0",
  };
}
