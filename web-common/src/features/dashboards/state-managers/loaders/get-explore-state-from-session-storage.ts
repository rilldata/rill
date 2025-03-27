import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { getExplorePresetForActivePage } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-active-page-store";
import {
  FromURLParamViewMap,
  ToActivePageViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import {
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

/**
 * Redirects to a view with params loaded from session storage.
 * 1. If no param is set then load the params for the default view from session storage.
 * 2. If only view param is set then load the params from session storage.
 * 3. If view=ttd and `measure` is the only other param set load from session storage.
 */
export function getExploreStateFromSessionStorage(
  exploreName: string,
  prefix: string | undefined,
  searchParams: URLSearchParams,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  defaultExplorePreset: V1ExplorePreset,
) {
  if (
    // exactly one param is set, but it is not `view`
    (searchParams.size === 1 &&
      !searchParams.has(ExploreStateURLParams.WebView)) ||
    // exactly 2 params are set and both `view` and `measure` are not set.
    (searchParams.size === 2 &&
      !searchParams.has(ExploreStateURLParams.WebView) &&
      !searchParams.has(ExploreStateURLParams.ExpandedMeasure)) ||
    // more than 2 params are set.
    searchParams.size > 2
  ) {
    return undefined;
  }

  const viewFromUrl = searchParams.get(ExploreStateURLParams.WebView) as string;
  const view = viewFromUrl
    ? FromURLParamViewMap[viewFromUrl]
    : (defaultExplorePreset.view ??
      V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED);
  const activePage = Number(ToActivePageViewMap[view] ?? "0");
  const exploreStateFromSessionStorage = getExplorePresetForActivePage(
    exploreName,
    prefix,
    activePage,
    metricsViewSpec,
    exploreSpec,
  );
  if (!exploreStateFromSessionStorage) {
    return undefined;
  }

  if (view === V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION) {
    exploreStateFromSessionStorage.tdd ??= {
      expandedMeasureName: "",
      chartType: TDDChart.DEFAULT,
      pinIndex: -1,
    };
    exploreStateFromSessionStorage.tdd.expandedMeasureName = searchParams.get(
      ExploreStateURLParams.ExpandedMeasure,
    ) as string;
  }

  return exploreStateFromSessionStorage;
}
