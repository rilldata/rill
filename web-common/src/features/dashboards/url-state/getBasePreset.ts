import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  URLStateDefaultTDDChartType,
  URLStateDefaultTimeRange,
  URLStateDefaultTimezone,
} from "@rilldata/web-common/features/dashboards/url-state/defaults";
import { ToURLParamTDDChartMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import type { LocalUserPreferences } from "@rilldata/web-common/features/dashboards/user-preferences";
import {
  V1ExploreComparisonMode,
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
} from "@rilldata/web-common/runtime-client";

export function getBasePreset(
  explore: V1ExploreSpec,
  preferences: LocalUserPreferences,
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

  return basePreset;
}
