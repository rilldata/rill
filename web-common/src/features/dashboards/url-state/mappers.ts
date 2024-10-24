import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1ExploreWebView,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";

export const FromURLParamViewMap: Record<string, V1ExploreWebView> = {
  overview: V1ExploreWebView.EXPLORE_ACTIVE_PAGE_OVERVIEW,
  pivot: V1ExploreWebView.EXPLORE_ACTIVE_PAGE_PIVOT,
  time_dimension: V1ExploreWebView.EXPLORE_ACTIVE_PAGE_TIME_DIMENSION,
};
export const ToURLParamViewMap = reverseMap(FromURLParamViewMap);

export const FromActivePageMap: Record<
  DashboardState_ActivePage,
  V1ExploreWebView
> = {
  [DashboardState_ActivePage.UNSPECIFIED]:
    V1ExploreWebView.EXPLORE_ACTIVE_PAGE_UNSPECIFIED,
  [DashboardState_ActivePage.DEFAULT]:
    V1ExploreWebView.EXPLORE_ACTIVE_PAGE_OVERVIEW,
  [DashboardState_ActivePage.PIVOT]: V1ExploreWebView.EXPLORE_ACTIVE_PAGE_PIVOT,
  [DashboardState_ActivePage.DIMENSION_TABLE]:
    V1ExploreWebView.EXPLORE_ACTIVE_PAGE_OVERVIEW,
  [DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL]:
    V1ExploreWebView.EXPLORE_ACTIVE_PAGE_TIME_DIMENSION,
};
export const ToActivePageViewMap = reverseMap(FromActivePageMap);
ToActivePageViewMap[V1ExploreWebView.EXPLORE_ACTIVE_PAGE_OVERVIEW] =
  DashboardState_ActivePage.DEFAULT;

export const FromURLParamTimeDimensionMap: Record<string, V1TimeGrain> = {
  "time.hour": V1TimeGrain.TIME_GRAIN_HOUR,
  "time.day": V1TimeGrain.TIME_GRAIN_DAY,
  "time.month": V1TimeGrain.TIME_GRAIN_MONTH,
};
export const ToURLParamTimeDimensionMap = reverseMap(
  FromURLParamTimeDimensionMap,
);

export const FromURLParamTDDChartMap: Record<string, TDDChart> = {
  timeseries: TDDChart.DEFAULT,
  bar: TDDChart.GROUPED_BAR,
  stacked_bar: TDDChart.STACKED_BAR,
  stacked_area: TDDChart.STACKED_AREA,
};
export const ToURLParamTDDChartMap = reverseMap(FromURLParamTDDChartMap);

export function reverseMap<
  K extends string | number,
  V extends string | number,
>(map: Partial<Record<K, V>>): Partial<Record<V, K>> {
  const revMap = {} as Partial<Record<V, K>>;
  for (const k in map) {
    revMap[map[k] as string | number] = k;
  }
  return revMap;
}
