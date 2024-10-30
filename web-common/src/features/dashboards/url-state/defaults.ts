import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import {
  V1ExploreComparisonMode,
  V1ExploreWebView,
} from "@rilldata/web-common/runtime-client";

export const URLStateDefaultView =
  V1ExploreWebView.EXPLORE_ACTIVE_PAGE_OVERVIEW;

export const URLStateDefaultTimeRange = "inf";
export const URLStateDefaultTimezone = "UTC";

export const URLStateDefaultCompareMode =
  V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE;

export const URLStateDefaultSortDirection = SortDirection.DESCENDING;

export const URLStateDefaultTDDChartType = TDDChart.DEFAULT;
export const ExplorePresetDefaultChartType = "timeseries";
