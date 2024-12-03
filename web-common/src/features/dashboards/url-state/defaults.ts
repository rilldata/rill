import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import {
  V1ExploreComparisonMode,
  V1ExploreWebView,
} from "@rilldata/web-common/runtime-client";

export const ExploreStateDefaultView =
  V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE;

export const ExploreStateDefaultTimeRange = "inf";
export const ExploreStateDefaultTimezone = "UTC";

export const ExploreStateDefaultCompareMode =
  V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE;

export const ExploreStateDefaultSortDirection = SortDirection.DESCENDING;

export const ExploreStateDefaultTDDChartType = TDDChart.DEFAULT;
export const ExploreStateDefaultChartType = "timeseries";
