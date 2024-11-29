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
  V1ExploreOverviewSortType,
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
  const defaultExplorePreset: V1ExplorePreset = {
    view: V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW,
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
    overviewSortType:
      V1ExploreOverviewSortType.EXPLORE_OVERVIEW_SORT_TYPE_VALUE,
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

  if (!defaultExplorePreset.timeGrain) {
    defaultExplorePreset.timeGrain = getDefaultPresetTimeGrain(
      defaultExplorePreset,
      fullTimeRange,
    );
  }

  if (defaultExplorePreset.comparisonMode) {
    Object.assign(
      defaultExplorePreset,
      getDefaultComparisonFields(defaultExplorePreset, explore),
    );
  }

  return defaultExplorePreset;
}

function getDefaultPresetTimeGrain(
  defaultExplorePreset: V1ExplorePreset,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
) {
  if (
    !defaultExplorePreset.timeRange ||
    !fullTimeRange?.timeRangeSummary?.min ||
    !fullTimeRange?.timeRangeSummary?.max
  )
    return "";

  const fullTimeStart = new Date(fullTimeRange.timeRangeSummary.min);
  const fullTimeEnd = new Date(fullTimeRange.timeRangeSummary.max);
  const timeRange = isoDurationToFullTimeRange(
    defaultExplorePreset.timeRange,
    fullTimeStart,
    fullTimeEnd,
    defaultExplorePreset.timezone,
  );

  return (
    ToURLParamTimeGrainMapMap[
      getDefaultTimeGrain(timeRange.start, timeRange.end)
    ] ?? ""
  );
}

function getDefaultComparisonFields(
  defaultExplorePreset: V1ExplorePreset,
  explore: V1ExploreSpec,
): V1ExplorePreset {
  if (
    defaultExplorePreset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_UNSPECIFIED
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

  // TODO: time comparison
  return {};
}