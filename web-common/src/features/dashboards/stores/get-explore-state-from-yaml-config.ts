import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { getMetricsViewTimeRangeFromExploreQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getGrainForRange } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { getValidComparisonOption } from "@rilldata/web-common/features/dashboards/time-controls/time-range-store";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { ToLegacySortTypeMap } from "@rilldata/web-common/features/dashboards/url-state/legacyMappers";
import { getExploreValidSpecQueryOptions } from "@rilldata/web-common/features/explores/selectors";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { arrayUnorderedEquals } from "@rilldata/web-common/lib/arrayUtils";
import { ISODurationToTimePreset } from "@rilldata/web-common/lib/time/ranges";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1ExploreComparisonMode,
  V1TimeGrain,
  type V1ExploreSpec,
  type V1TimeRangeSummary,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";
import {
  DateTimeUnitToV1TimeGrain,
  isGrainAllowed,
} from "@rilldata/web-common/lib/time/new-grains";
import {
  createAndExpression,
  flattenExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";

export function getExploreStateFromYAMLConfig(
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
  smallestTimeGrain: V1TimeGrain | undefined = undefined,
) {
  // TODO: support all fields from V1ExplorePreset. Not urgent since we do not parse them in backend.
  return <Partial<ExploreState>>{
    activePage: DashboardState_ActivePage.DEFAULT,

    ...getExploreFilterStateFromYAMLConfig(exploreSpec),
    ...getExploreTimeStateFromYAMLConfig(
      exploreSpec,
      timeRangeSummary,
      smallestTimeGrain,
    ),
    ...getExploreViewStateFromYAMLConfig(exploreSpec),
  };
}

export function createUrlForExploreYAMLDefaultState(
  client: RuntimeClient,
  exploreNameStore: Readable<string>,
) {
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(client, exploreNameStore),
  );
  const timeRangeQuery = createQuery(
    getMetricsViewTimeRangeFromExploreQueryOptions(client, exploreNameStore),
  );

  return derived(
    [validSpecQuery, timeRangeQuery],
    ([validSpecResp, timeRangeResp]) => {
      const metricsViewSpec = validSpecResp.data?.metricsViewSpec ?? {};
      const exploreSpec = validSpecResp.data?.exploreSpec ?? {};
      const timeRangeSummary = timeRangeResp.data?.timeRangeSummary;

      const exploreStateFromYAMLConfig = getExploreStateFromYAMLConfig(
        exploreSpec,
        timeRangeSummary,
        metricsViewSpec.smallestTimeGrain,
      );

      const timeControlState = getTimeControlState(
        metricsViewSpec,
        exploreSpec,
        timeRangeSummary,
        exploreStateFromYAMLConfig as ExploreState,
      );

      const urlParams = convertPartialExploreStateToUrlParams(
        exploreSpec,
        metricsViewSpec,
        exploreStateFromYAMLConfig,
        timeControlState,
      );
      return `?${urlParams.toString()}`;
    },
  );
}

export function getExploreFilterStateFromYAMLConfig(
  exploreSpec: V1ExploreSpec,
): Partial<ExploreState> {
  const filter = exploreSpec.defaultPreset?.where;
  if (!filter && !exploreSpec.defaultPreset?.pinnedFilters?.length) {
    return {};
  }

  const flattened = filter
    ? flattenExpression(filter)
    : createAndExpression([]);

  const { dimensionThresholdFilters, dimensionFilters } =
    splitWhereFilter(flattened);

  // Build dimensionFilterExcludeMode from the parsed filter expressions.
  // NIN (NOT IN) operations indicate exclude mode for that dimension.
  const dimensionFilterExcludeMode = new Map<string, boolean>();
  for (const expr of dimensionFilters.cond?.exprs ?? []) {
    const ident = expr.cond?.exprs?.[0]?.ident;
    if (ident && expr.cond?.op === V1Operation.OPERATION_NIN) {
      dimensionFilterExcludeMode.set(ident, true);
    }
  }

  return {
    whereFilter: dimensionFilters,
    dimensionThresholdFilters,
    dimensionFilterExcludeMode,
    pinnedFilters: new Set(exploreSpec.defaultPreset?.pinnedFilters ?? []),
  };
}

function getExploreTimeStateFromYAMLConfig(
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
  smallestTimeGrain: V1TimeGrain | undefined = undefined,
): Partial<ExploreState> {
  const exploreTimeState: Partial<ExploreState> = {};
  if (!exploreSpec.defaultPreset || !timeRangeSummary) {
    return exploreTimeState;
  }
  const defaultPreset = exploreSpec.defaultPreset;

  if (defaultPreset.timeRange) {
    exploreTimeState.selectedTimeRange = {
      name: defaultPreset.timeRange,
    } as DashboardTimeControls;

    if (defaultPreset.timeGrain) {
      exploreTimeState.selectedTimeRange.interval =
        DateTimeUnitToV1TimeGrain[defaultPreset.timeGrain];
    } else {
      const grainForRange = getGrainForRange(
        defaultPreset.timeRange,
        defaultPreset.timezone,
        timeRangeSummary,
      );

      const interval = isGrainAllowed(grainForRange, smallestTimeGrain)
        ? grainForRange
        : smallestTimeGrain;

      exploreTimeState.selectedTimeRange.interval = interval;
    }
  }

  if (defaultPreset.timezone) {
    exploreTimeState.selectedTimezone = defaultPreset.timezone;
  }

  switch (defaultPreset.comparisonMode) {
    case V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME: {
      exploreTimeState.showTimeComparison = true;
      let comparisonTimeRangeName = defaultPreset.compareTimeRange;
      if (!comparisonTimeRangeName) {
        comparisonTimeRangeName = getDefaultComparisonTimeRangeName(
          exploreSpec,
          defaultPreset.timeRange ?? TimeRangePreset.LAST_12_MONTHS,
          defaultPreset.timezone,
          timeRangeSummary,
        );
      }
      exploreTimeState.selectedComparisonTimeRange = {
        name: comparisonTimeRangeName,
      } as DashboardTimeControls;
      break;
    }

    case V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION:
      exploreTimeState.selectedComparisonDimension =
        defaultPreset.comparisonDimension || exploreSpec.dimensions?.[0];
  }

  return exploreTimeState;
}

function getDefaultComparisonTimeRangeName(
  exploreSpec: V1ExploreSpec,
  timeRangeName: string,
  timezone: string | undefined,
  timeRangeSummary: V1TimeRangeSummary,
) {
  const timePreset = ISODurationToTimePreset(timeRangeName, true);
  if (!timePreset) return undefined;

  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(timeRangeSummary.min!),
    end: new Date(timeRangeSummary.max!),
  };

  const timeRange = isoDurationToFullTimeRange(
    timePreset,
    allTimeRange.start,
    allTimeRange.end,
    timezone,
  );

  const comparisonTimeRangeName = getValidComparisonOption(
    exploreSpec.timeRanges,
    timeRange,
    undefined,
    allTimeRange,
    timezone,
  );

  return comparisonTimeRangeName;
}

function getExploreViewStateFromYAMLConfig(
  exploreSpec: V1ExploreSpec,
): Partial<ExploreState> {
  const exploreViewState: Partial<ExploreState> = {};
  if (!exploreSpec.defaultPreset) return exploreViewState;
  const defaultPreset = exploreSpec.defaultPreset;

  if (defaultPreset.measures) {
    exploreViewState.visibleMeasures = defaultPreset.measures;
    exploreViewState.allMeasuresVisible = arrayUnorderedEquals(
      defaultPreset.measures,
      exploreSpec.measures ?? [],
    );
  }

  if (defaultPreset.dimensions) {
    exploreViewState.visibleDimensions = defaultPreset.dimensions;
    exploreViewState.allDimensionsVisible = arrayUnorderedEquals(
      defaultPreset.dimensions,
      exploreSpec.dimensions ?? [],
    );
  }

  if (defaultPreset.exploreSortBy) {
    exploreViewState.leaderboardSortByMeasureName = defaultPreset.exploreSortBy;
  } else if (exploreViewState.visibleMeasures?.length) {
    exploreViewState.leaderboardSortByMeasureName =
      exploreViewState.visibleMeasures[0];
  }

  if ("exploreSortAsc" in defaultPreset) {
    exploreViewState.sortDirection = defaultPreset.exploreSortAsc
      ? SortDirection.ASCENDING
      : SortDirection.DESCENDING;
  }

  if (defaultPreset.exploreSortType) {
    exploreViewState.dashboardSortType = Number(
      ToLegacySortTypeMap[defaultPreset.exploreSortType],
    );
  }

  if (defaultPreset.exploreLeaderboardMeasures?.length) {
    exploreViewState.leaderboardMeasureNames =
      defaultPreset.exploreLeaderboardMeasures;
  } else if (exploreViewState.leaderboardSortByMeasureName) {
    exploreViewState.leaderboardMeasureNames = [
      exploreViewState.leaderboardSortByMeasureName,
    ];
  }

  if (defaultPreset.exploreLeaderboardShowContextForAllMeasures !== undefined) {
    exploreViewState.leaderboardShowContextForAllMeasures =
      defaultPreset.exploreLeaderboardShowContextForAllMeasures;
  }

  if (defaultPreset.exploreExpandedDimension) {
    exploreViewState.selectedDimensionName =
      defaultPreset.exploreExpandedDimension;
    exploreViewState.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
  }

  return exploreViewState;
}
