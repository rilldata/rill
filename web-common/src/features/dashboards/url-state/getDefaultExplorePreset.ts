import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  getDefaultTimeRange,
  getDefaultTimeZone,
} from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import { getValidComparisonOption } from "@rilldata/web-common/features/dashboards/time-controls/time-range-store";
import { getDefaultTimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { ToURLParamTDDChartMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ISODurationToTimePreset } from "@rilldata/web-common/lib/time/ranges";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  V1ExploreComparisonMode,
  V1ExploreSortType,
  V1ExploreWebView,
  type V1ExplorePreset,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { ALL_TIME_RANGE_ALIAS } from "../time-controls/new-time-controls";
import { DEFAULT_TIMEZONES } from "@rilldata/web-common/lib/time/config";
import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";

export function getDefaultExplorePreset(
  explore: V1ExploreSpec,
  metricsViewSpec: V1MetricsViewSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  const defaultMeasure =
    explore.defaultPreset?.measures?.[0] ?? explore.measures?.[0];

  const defaultExplorePreset: V1ExplorePreset = {
    view: V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
    where: createAndExpression([]),
    dimensionsWithInlistFilter: [],

    measures: explore.measures,
    dimensions: explore.dimensions,

    timeRange:
      explore.defaultPreset?.timeRange ||
      getDefaultTimeRange(metricsViewSpec.smallestTimeGrain, timeRangeSummary),
    timezone: getDefaultTimeZone(explore),
    timeGrain: "",
    comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE,
    compareTimeRange: "",
    comparisonDimension: "",

    exploreSortBy: defaultMeasure,
    exploreSortAsc: false,
    exploreSortType: V1ExploreSortType.EXPLORE_SORT_TYPE_VALUE,
    exploreExpandedDimension: "",
    exploreLeaderboardMeasures: defaultMeasure ? [defaultMeasure] : [],
    exploreLeaderboardShowContextForAllMeasures: false,

    timeDimensionMeasure: "",
    timeDimensionChartType: ToURLParamTDDChartMap[TDDChart.DEFAULT],
    timeDimensionPin: false,

    pivotCols: [],
    pivotRows: [],
    pivotSortBy: "",
    pivotSortAsc: false,
    pivotTableMode: "nest",

    ...(explore.defaultPreset ?? {}),
  };

  if (!defaultExplorePreset.timeGrain) {
    defaultExplorePreset.timeGrain = getDefaultPresetTimeGrain(
      defaultExplorePreset,
      timeRangeSummary,
    );
  }

  if (defaultExplorePreset.comparisonMode) {
    Object.assign(
      defaultExplorePreset,
      getDefaultComparisonFields(
        defaultExplorePreset,
        explore,
        timeRangeSummary,
      ),
    );
  }

  return defaultExplorePreset;
}

function getDefaultPresetTimeGrain(
  defaultExplorePreset: V1ExplorePreset,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  if (
    !defaultExplorePreset.timeRange ||
    !timeRangeSummary?.min ||
    !timeRangeSummary?.max
  )
    return "";

  const fullTimeStart = new Date(timeRangeSummary.min);
  const fullTimeEnd = new Date(timeRangeSummary.max);
  const timeRange = isoDurationToFullTimeRange(
    defaultExplorePreset.timeRange,
    fullTimeStart,
    fullTimeEnd,
    defaultExplorePreset.timezone,
  );

  return (
    V1TimeGrainToDateTimeUnit[
      getDefaultTimeGrain(timeRange.start, timeRange.end)
    ] ?? ""
  );
}

function getDefaultComparisonFields(
  defaultExplorePreset: V1ExplorePreset,
  explore: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
): V1ExplorePreset {
  if (
    defaultExplorePreset.comparisonMode ===
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_UNSPECIFIED ||
    defaultExplorePreset.comparisonMode ===
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE
  ) {
    return {};
  }

  if (
    defaultExplorePreset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION
  ) {
    return {
      comparisonDimension:
        defaultExplorePreset.comparisonDimension || explore.dimensions?.[0],
    };
  }

  if (
    !defaultExplorePreset.timeRange ||
    defaultExplorePreset.timeRange === ALL_TIME_RANGE_ALIAS ||
    !timeRangeSummary?.min ||
    !timeRangeSummary?.max
  ) {
    return {};
  }

  let comparisonOption = defaultExplorePreset.compareTimeRange;

  if (!comparisonOption) {
    const preset = ISODurationToTimePreset(
      defaultExplorePreset.timeRange,
      true,
    );
    if (!preset) return {};

    const allTimeRange = {
      name: TimeRangePreset.ALL_TIME,
      start: new Date(timeRangeSummary.min),
      end: new Date(timeRangeSummary.max),
    };

    const timeRange = isoDurationToFullTimeRange(
      preset,
      allTimeRange.start,
      allTimeRange.end,
      defaultExplorePreset.timezone,
    );

    comparisonOption = getValidComparisonOption(
      explore.timeRanges,
      timeRange,
      undefined,
      allTimeRange,
    );
  }

  return {
    compareTimeRange: comparisonOption,
    exploreSortType:
      defaultExplorePreset.exploreSortType ??
      V1ExploreSortType.EXPLORE_SORT_TYPE_DELTA_PERCENT,
  };
}

export function getPinnedTimeZones(explore: V1ExploreSpec) {
  const yamlTimeZones = explore.timeZones;
  if (!yamlTimeZones || !yamlTimeZones.length) return DEFAULT_TIMEZONES;
  return yamlTimeZones;
}
