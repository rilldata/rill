import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { getDefaultTimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
import {
  URLStateDefaultTDDChartType,
  URLStateDefaultTimeRange,
  URLStateDefaultTimezone,
} from "@rilldata/web-common/features/dashboards/url-state/defaults";
import {
  ToURLParamTDDChartMap,
  ToURLParamTimeGrainMapMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import type { LocalUserPreferences } from "@rilldata/web-common/features/dashboards/user-preferences";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  V1ExploreComparisonMode,
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
  type V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";

export function getBasePreset(
  explore: V1ExploreSpec,
  preferences: LocalUserPreferences,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  const basePreset: V1ExplorePreset = {
    view: V1ExploreWebView.EXPLORE_ACTIVE_PAGE_OVERVIEW,
    where: createAndExpression([]),

    measures: explore.measures,
    dimensions: explore.dimensions,

    timeRange: URLStateDefaultTimeRange,
    timezone: preferences.timeZone ?? URLStateDefaultTimezone,
    timeGrain: "",
    comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE,
    compareTimeRange: "",
    comparisonDimension: "",

    overviewSortBy: explore.measures?.[0],
    overviewSortAsc: false,
    overviewExpandedDimension: "",

    timeDimensionMeasure: "",
    timeDimensionChartType: ToURLParamTDDChartMap[URLStateDefaultTDDChartType],
    timeDimensionPin: false,

    pivotCols: [],
    pivotRows: [],
    pivotSortBy: "",
    pivotSortAsc: false,

    ...(explore.defaultPreset ?? {}),
  };

  if (!basePreset.timeGrain) {
    basePreset.timeGrain = getDefaultPresetTimeGrain(basePreset, fullTimeRange);
  }

  if (basePreset.comparisonMode) {
    Object.assign(basePreset, getDefaultComparisonFields(basePreset, explore));
  }

  return basePreset;
}

function getDefaultPresetTimeGrain(
  basePreset: V1ExplorePreset,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  if (
    !basePreset.timeRange ||
    !fullTimeRange?.timeRangeSummary?.min ||
    !fullTimeRange?.timeRangeSummary?.max
  )
    return "";

  const fullTimeStart = new Date(fullTimeRange.timeRangeSummary.min);
  const fullTimeEnd = new Date(fullTimeRange.timeRangeSummary.max);
  const timeRange = isoDurationToFullTimeRange(
    basePreset.timeRange,
    fullTimeStart,
    fullTimeEnd,
    basePreset.timezone,
  );

  return (
    ToURLParamTimeGrainMapMap[
      getDefaultTimeGrain(timeRange.start, timeRange.end)
    ] ?? ""
  );
}

function getDefaultComparisonFields(
  basePreset: V1ExplorePreset,
  explore: V1ExploreSpec,
): V1ExplorePreset {
  if (
    basePreset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_UNSPECIFIED
  ) {
    return {};
  }

  if (
    basePreset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION
  ) {
    return {
      comparisonDimension:
        basePreset.comparisonDimension || explore.dimensions?.[0],
    };
  }

  // TODO: time comparison
  return {};
}
