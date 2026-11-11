import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { getMetricsViewTimeRangeFromExploreQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getGrainForRange } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import {
  type ExploreFieldNames,
  getExploreFieldNames,
} from "@rilldata/web-common/features/dashboards/stores/validate-and-clean-explore-state";
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
  type V1MetricsViewSpec,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";
import {
  getAvailableTimeGrains,
  TokenToV1TimeGrain,
} from "@rilldata/web-common/lib/time/new-grains";

export function getExploreStateFromYAMLConfig(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  // Only fields the metrics view still defines can be queried. See getExploreFieldNames.
  const exploreFields = getExploreFieldNames(metricsViewSpec, exploreSpec);

  // TODO: support all fields from V1ExplorePreset. Not urgent since we do not parse them in backend.
  return <Partial<ExploreState>>{
    activePage: DashboardState_ActivePage.DEFAULT,

    ...getExploreTimeStateFromYAMLConfig(
      exploreSpec,
      exploreFields,
      timeRangeSummary,
      metricsViewSpec.smallestTimeGrain,
      metricsViewSpec.timeGrains,
    ),
    ...getExploreViewStateFromYAMLConfig(exploreSpec, exploreFields),
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
        metricsViewSpec,
        exploreSpec,
        timeRangeSummary,
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

function getExploreTimeStateFromYAMLConfig(
  exploreSpec: V1ExploreSpec,
  exploreFields: ExploreFieldNames,
  timeRangeSummary: V1TimeRangeSummary | undefined,
  smallestTimeGrain: V1TimeGrain | undefined,
  timeGrains: V1TimeGrain[] | undefined,
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
        TokenToV1TimeGrain[defaultPreset.timeGrain];
      exploreTimeState.adaptiveTimeGrain = false;
    } else {
      const grainForRange = getGrainForRange(
        defaultPreset.timeRange,
        defaultPreset.timezone,
        timeRangeSummary,
      );

      const availableTimeGrains = getAvailableTimeGrains({
        smallestTimeGrain,
        timeGrains,
      });
      const interval =
        grainForRange && availableTimeGrains.includes(grainForRange)
          ? grainForRange
          : (availableTimeGrains[0] ?? smallestTimeGrain);

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
        defaultPreset.comparisonDimension &&
        exploreFields.dimensions.includes(defaultPreset.comparisonDimension)
          ? defaultPreset.comparisonDimension
          : exploreFields.dimensions[0];
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
  exploreFields: ExploreFieldNames,
): Partial<ExploreState> {
  const exploreViewState: Partial<ExploreState> = {};
  if (!exploreSpec.defaultPreset) return exploreViewState;
  const defaultPreset = exploreSpec.defaultPreset;
  const measures = new Set(exploreFields.measures);
  const dimensions = new Set(exploreFields.dimensions);

  // Fields are left unset when the defaults name none that still exist,
  // so that the cascading merge fills them from the next source.
  const visibleMeasures = defaultPreset.measures?.filter((m) =>
    measures.has(m),
  );
  if (visibleMeasures?.length) {
    exploreViewState.visibleMeasures = visibleMeasures;
    exploreViewState.allMeasuresVisible = arrayUnorderedEquals(
      visibleMeasures,
      exploreFields.measures,
    );
  }

  const visibleDimensions = defaultPreset.dimensions?.filter((d) =>
    dimensions.has(d),
  );
  if (visibleDimensions?.length) {
    exploreViewState.visibleDimensions = visibleDimensions;
    exploreViewState.allDimensionsVisible = arrayUnorderedEquals(
      visibleDimensions,
      exploreFields.dimensions,
    );
  }

  if (
    defaultPreset.exploreSortBy &&
    measures.has(defaultPreset.exploreSortBy)
  ) {
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

  const leaderboardMeasures = defaultPreset.exploreLeaderboardMeasures?.filter(
    (m) => measures.has(m),
  );
  if (leaderboardMeasures?.length) {
    exploreViewState.leaderboardMeasureNames = leaderboardMeasures;
  } else if (exploreViewState.leaderboardSortByMeasureName) {
    exploreViewState.leaderboardMeasureNames = [
      exploreViewState.leaderboardSortByMeasureName,
    ];
  }

  if (defaultPreset.exploreLeaderboardShowContextForAllMeasures !== undefined) {
    exploreViewState.leaderboardShowContextForAllMeasures =
      defaultPreset.exploreLeaderboardShowContextForAllMeasures;
  }

  if (
    defaultPreset.exploreExpandedDimension &&
    dimensions.has(defaultPreset.exploreExpandedDimension)
  ) {
    exploreViewState.selectedDimensionName =
      defaultPreset.exploreExpandedDimension;
    exploreViewState.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
  }

  return exploreViewState;
}
