import {
  applyExactMeasures,
  mapEphemeralMeasuresForRequest,
} from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
import { requestExactNames } from "@rilldata/web-common/features/dashboards/exact-measures/exact-measures";
import { exploreExactSurface } from "@rilldata/web-common/features/dashboards/exact-measures/explore-surfaces";
import { getComparisonRequestMeasures } from "@rilldata/web-common/features/dashboards/dashboard-utils";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { getDimensionQueryMeasuresAndSort } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  mapSelectedComparisonTimeRangeToV1TimeRange,
  mapSelectedTimeRangeToV1TimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1Expression,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationRequest,
  V1MetricsViewSpec,
  V1Query,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { buildWhereParamForDimensionTableAndTDDExports } from "../../exports/export-filters";
import { dimensionSearchText as dimensionSearchTextStore } from "../stores/dashboard-stores";

export function getDimensionTableExportQuery(
  ctx: StateManagers,
  isScheduled: boolean,
): V1Query | undefined {
  const metricsViewName = get(ctx.metricsViewName);
  const exploreState = get(ctx.dashboardStore);
  const timeControlState = get(useTimeControlStore(ctx));
  const validSpecStore = get(ctx.validSpecStore);
  const dimensionSearchText = get(dimensionSearchTextStore);

  if (
    !validSpecStore.data?.explore ||
    !validSpecStore.data.metricsView ||
    !timeControlState.ready
  )
    return undefined;

  let timeRange: V1TimeRange | undefined;
  if (isScheduled) {
    timeRange = mapSelectedTimeRangeToV1TimeRange(
      timeControlState.selectedTimeRange,
      exploreState.selectedTimezone,
      validSpecStore.data.explore,
    );
  } else {
    timeRange = {
      start: timeControlState.timeStart,
      end: timeControlState.timeEnd,
    };
  }
  if (!timeRange) return undefined;

  let comparisonTimeRange: V1TimeRange | undefined;
  if (timeControlState.showTimeComparison) {
    if (isScheduled) {
      comparisonTimeRange = mapSelectedComparisonTimeRangeToV1TimeRange(
        timeControlState.selectedComparisonTimeRange,
        timeControlState.showTimeComparison,
        timeRange,
      );
    } else {
      comparisonTimeRange = {
        start: timeControlState.comparisonTimeStart,
        end: timeControlState.comparisonTimeEnd,
      };
    }
  }

  // A one-shot export downloads what is on screen, so it carries the viewer's
  // exact selection. A scheduled report must not: exact values are only
  // fetched after an explicit click, never on a schedule.
  const exactMeasures = isScheduled
    ? undefined
    : requestExactNames(
        [
          ...exploreState.visibleMeasures,
          exploreState.leaderboardSortByMeasureName,
        ],
        get(ctx.selectors.exactMeasures.exactMeasureContext)(
          exploreExactSurface.dimensionTable,
        ),
      );

  const query: V1Query = {
    metricsViewAggregationRequest: getDimensionTableAggregationRequestForTime({
      instanceId: ctx.runtimeClient.instanceId,
      metricsViewName,
      exploreState,
      metricsViewSpec: validSpecStore.data.metricsView,
      filter: ctx.expressionFilterManager.exprByMetricsView[metricsViewName],
      timeRange,
      comparisonTimeRange,
      dimensionSearchText,
      exactMeasures,
    }),
  };

  return query;
}

export function getDimensionTableAggregationRequestForTime({
  instanceId,
  metricsViewName,
  exploreState,
  metricsViewSpec,
  filter,
  timeRange,
  comparisonTimeRange,
  dimensionSearchText,
  exactMeasures,
}: {
  instanceId: string;
  metricsViewName: string;
  exploreState: ExploreState;
  metricsViewSpec: V1MetricsViewSpec;
  filter?: V1Expression | undefined;
  timeRange: V1TimeRange;
  comparisonTimeRange: V1TimeRange | undefined;
  dimensionSearchText: string;
  // Measures to request with `exact: true`; see `requestExactNames`.
  exactMeasures?: ReadonlySet<string>;
}): V1MetricsViewAggregationRequest {
  const measures: V1MetricsViewAggregationMeasure[] =
    exploreState.visibleMeasures.map((name) => ({
      name: name,
    }));

  let apiSortName = exploreState.leaderboardSortByMeasureName;
  if (!exploreState.visibleMeasures.includes(apiSortName)) {
    // if selected sort measure is not visible add it to list
    measures.push({ name: apiSortName });
  }
  if (comparisonTimeRange) {
    // insert beside the correct measure
    measures.splice(
      measures.findIndex((m) => m.name === apiSortName) + 1,
      0,
      ...getComparisonRequestMeasures(apiSortName),
    );
    switch (exploreState.dashboardSortType) {
      case DashboardState_LeaderboardSortType.DELTA_ABSOLUTE:
        apiSortName += ComparisonDeltaAbsoluteSuffix;
        break;
      case DashboardState_LeaderboardSortType.DELTA_PERCENT:
        apiSortName += ComparisonDeltaRelativeSuffix;
        break;
    }
  }

  // The query groups only by the selected dimension, as in the dimension table itself.
  const { measures: supportedMeasures, sort } =
    getDimensionQueryMeasuresAndSort(
      exploreState,
      metricsViewSpec,
      measures,
      exploreState.selectedDimensionName!,
      apiSortName,
    );

  const where = buildWhereParamForDimensionTableAndTDDExports(
    filter,
    exploreState.selectedDimensionName!, // must exist when viewing a dimension table
    dimensionSearchText,
  );

  return {
    instanceId,
    metricsView: metricsViewName,
    dimensions: [
      {
        name: exploreState.selectedDimensionName,
      },
    ],
    measures: applyExactMeasures(
      mapEphemeralMeasuresForRequest(
        supportedMeasures,
        exploreState.ephemeralMeasures,
      ),
      exactMeasures,
    ),
    timeRange,
    ...(comparisonTimeRange ? { comparisonTimeRange } : {}),
    sort,
    where,
    offset: "0",
  };
}
