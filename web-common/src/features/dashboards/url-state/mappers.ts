import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { reverseMap } from "@rilldata/web-common/lib/map-utils.ts";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1ExploreSortType,
  V1ExploreWebView,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";

export enum ExploreUrlWebView {
  Explore = "explore",
  Pivot = "pivot",
  TimeDimension = "tdd",
}
export const FromURLParamViewMap: Record<ExploreUrlWebView, V1ExploreWebView> =
  {
    explore: V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
    pivot: V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
    tdd: V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
  };
export const ToURLParamViewMap = reverseMap(FromURLParamViewMap);

export const FromActivePageMap: Record<
  DashboardState_ActivePage,
  V1ExploreWebView
> = {
  [DashboardState_ActivePage.UNSPECIFIED]:
    V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED,
  [DashboardState_ActivePage.DEFAULT]:
    V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
  [DashboardState_ActivePage.PIVOT]: V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
  [DashboardState_ActivePage.DIMENSION_TABLE]:
    V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE,
  [DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL]:
    V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
};
export const ToActivePageViewMap = reverseMap(FromActivePageMap);
ToActivePageViewMap[V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE] =
  DashboardState_ActivePage.DEFAULT;

export const FromURLParamsSortTypeMap: Record<string, V1ExploreSortType> = {
  value: V1ExploreSortType.EXPLORE_SORT_TYPE_VALUE,
  percent: V1ExploreSortType.EXPLORE_SORT_TYPE_PERCENT,
  delta_abs: V1ExploreSortType.EXPLORE_SORT_TYPE_DELTA_ABSOLUTE,
  delta_percent: V1ExploreSortType.EXPLORE_SORT_TYPE_DELTA_PERCENT,
  dim: V1ExploreSortType.EXPLORE_SORT_TYPE_DIMENSION,
};
export const ToURLParamSortTypeMap = reverseMap(FromURLParamsSortTypeMap);

export const FromURLParamTimeGrainMap: Record<string, V1TimeGrain> = {};
Object.values(TIME_GRAIN).forEach((tg) => {
  FromURLParamTimeGrainMap[tg.label] = tg.grain;
});
export const ToURLParamTimeGrainMapMap = reverseMap(FromURLParamTimeGrainMap);

export const FromURLParamTimeDimensionMap: Record<string, V1TimeGrain> = {};
Object.values(TIME_GRAIN).forEach((tg) => {
  FromURLParamTimeDimensionMap["time." + tg.label] = tg.grain;
});
export const ToURLParamTimeDimensionMap = reverseMap(
  FromURLParamTimeDimensionMap,
);

export const FromURLParamTDDChartMap: Record<string, TDDChart> = {
  timeseries: TDDChart.DEFAULT, // Backwards compatibility, this was default value when we 1st did this feature
  line: TDDChart.DEFAULT,
  bar: TDDChart.GROUPED_BAR,
  stacked_bar: TDDChart.STACKED_BAR,
  stacked_area: TDDChart.STACKED_AREA,
};
export const ToURLParamTDDChartMap = reverseMap(FromURLParamTDDChartMap);

export const FromURLParamTimeRangePresetMap: Record<string, TimeRangePreset> =
  {};
Object.keys(TimeRangePreset).forEach(
  (tr) =>
    (FromURLParamTimeRangePresetMap[TimeRangePreset[tr]] =
      tr as TimeRangePreset),
);
