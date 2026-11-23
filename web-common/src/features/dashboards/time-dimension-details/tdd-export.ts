import {
  applyExactMeasures,
  mapEphemeralMeasuresForRequest,
} from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
import { requestExactNames } from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
import { exploreExactSurface } from "@rilldata/web-common/features/dashboards/exact-measures/explore-surfaces";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { mapTimeManagerRangeToV1TimeRange } from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { getTimeDimensionForComparison } from "@rilldata/web-common/features/dashboards/time-series/utils";
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
  const validSpec = get(ctx.validSpecStore);
  const dimensionSearchText = get(dimensionSearchTextStore);
  const timeFilterManager = ctx.timeFilterManager;

  let timeRange: V1TimeRange | undefined;
  if (timeFilterManager.ready) {
    timeRange = isScheduled
      ? mapTimeManagerRangeToV1TimeRange(timeFilterManager)
      : { start: timeFilterManager.timeStart, end: timeFilterManager.timeEnd };
  }

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
      timeRange,
      metricsViewSpec: validSpec.data?.metricsView,
      exploreSpec: validSpec.data?.explore,
      dimensionSearchText,
      exactMeasures,
    }),
  };

  return query;
}

export function getTDDAggregationRequest({
  instanceId,
  metricsViewName,
  exploreState,
  timeRange,
  metricsViewSpec,
  exploreSpec,
  dimensionSearchText,
  exactMeasures,
}: {
  instanceId: string;
  metricsViewName: string;
  exploreState: ExploreState;
  // Undefined until the time filters are ready.
  timeRange: V1TimeRange | undefined;
  metricsViewSpec: V1MetricsViewSpec | undefined;
  exploreSpec: V1ExploreSpec | undefined;
  dimensionSearchText: string;
  // Measures to request with `exact: true`; see `requestExactNames`.
  exactMeasures?: ReadonlySet<string>;
}): undefined | V1MetricsViewAggregationRequest {
  if (
    !metricsViewSpec ||
    !exploreSpec ||
    !timeRange ||
    !exploreState.tdd.expandedMeasureName
  )
    return undefined;

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

  const { dimension: timeDimensionRequest, field: timeField } =
    getTimeDimensionForComparison(
      dimensionName,
      timeDimension,
      exploreState.selectedTimeRange?.interval,
      exploreState.selectedTimezone,
    );

  return {
    instanceId,
    metricsView: metricsViewName,
    dimensions: [{ name: dimensionName }, timeDimensionRequest],
    measures,
    timeRange,
    pivotOn: [timeField],
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
