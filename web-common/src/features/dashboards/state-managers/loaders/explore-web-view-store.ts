import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import {
  ExploreUrlWebView,
  FromActivePageMap,
  FromURLParamViewMap,
  ToURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import {
  type V1ExploreSpec,
  V1ExploreWebView,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { copyUrlSearchParamsForView } from "web-common/src/features/dashboards/url-state/explore-web-view-specific-url-params";

export function getKeyForSessionStore(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  webView: ExploreUrlWebView,
) {
  return `rill:app:explore:${storageNamespacePrefix ?? ""}${exploreName}:${webView}`.toLowerCase();
}

/**
 * Save the current explore state as "most recent explore" state in sessionStorage.
 * Makes sure to update all views with global and common fields.
 * Stores in url params format so that the converter code is shared.
 */
export function updateExploreSessionStore(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  exploreState: ExploreState,
  exploreSpec: V1ExploreSpec,
  metricsViewSpec: V1MetricsViewSpec,
  timeControlsState: TimeControlState | undefined,
) {
  const apiWebView =
    FromActivePageMap[exploreState.activePage] ??
    V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE;
  const curWebView = ToURLParamViewMap[apiWebView];
  // type guard
  if (!curWebView) return;

  // Build the url search params for the entire state
  const urlSearchParams = convertPartialExploreStateToUrlParams(
    exploreSpec,
    metricsViewSpec,
    exploreState,
    timeControlsState,
  );
  try {
    // Store the full url for the web view
    setExploreStateForWebView(
      exploreName,
      storageNamespacePrefix,
      curWebView,
      urlSearchParams.toString(),
    );
  } catch {
    // no-op
  }

  // We need to make sure other views are in sync
  for (const otherWebViewStr in FromURLParamViewMap) {
    const otherWebView = otherWebViewStr as ExploreUrlWebView;
    if (otherWebView === curWebView) continue;

    const otherWebViewKey = getKeyForSessionStore(
      exploreName,
      storageNamespacePrefix,
      otherWebView,
    );
    const otherWebViewRawSearch = sessionStorage.getItem(otherWebViewKey) ?? "";
    const otherWebViewUrlParams = new URLSearchParams(otherWebViewRawSearch);

    // Copy relevant params from current view to the other view
    copyUrlSearchParamsForView(
      curWebView,
      urlSearchParams,
      otherWebView,
      otherWebViewUrlParams,
    );

    try {
      // Store the full url for the other web view
      setExploreStateForWebView(
        exploreName,
        storageNamespacePrefix,
        otherWebView,
        otherWebViewUrlParams.toString(),
      );
    } catch {
      // no-op
    }
  }
}

/**
 * Returns explore state filled with extra fields stored when user last visited a particular view.
 * 1. If no param is set then load the params for the default view from session storage.
 * 2. If only view param is set then load the params from session storage.
 * 3. If view=ttd and `measure` is the only other param set load from session storage.
 */
export function getPartialExploreStateFromSessionStorage(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  searchParams: URLSearchParams,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
) {
  if (shouldSkipSessionStorage(searchParams)) {
    return undefined;
  }

  const viewFromUrl = (searchParams.get(ExploreStateURLParams.WebView) ??
    ExploreUrlWebView.Explore) as ExploreUrlWebView;
  const key = getKeyForSessionStore(
    exploreName,
    storageNamespacePrefix,
    viewFromUrl,
  );

  try {
    const storedUrlSearch = sessionStorage.getItem(key);
    if (!storedUrlSearch) return undefined;
    const storedUrlSearchParams = new URLSearchParams(storedUrlSearch);

    const { partialExploreState: storedExploreState } =
      convertURLSearchParamsToExploreState(
        storedUrlSearchParams,
        metricsViewSpec,
        exploreSpec,
        {},
      );

    // TDD is different from other views. It has a variable that is expanded measure.
    // So we need to copy over the actual measure from current url but keep other params.
    if (viewFromUrl === ExploreUrlWebView.TimeDimension) {
      // type safety
      storedExploreState.tdd ??= {
        expandedMeasureName: "",
        chartType: TDDChart.DEFAULT,
        pinIndex: -1,
      };
      // copy over the expanded measure from current url search params.
      storedExploreState.tdd.expandedMeasureName = searchParams.get(
        ExploreStateURLParams.ExpandedMeasure,
      ) as string;
    }

    return storedExploreState;
  } catch {
    return undefined;
  }
}

export function clearExploreSessionStore(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
) {
  for (const otherWebView in FromURLParamViewMap) {
    const key = getKeyForSessionStore(
      exploreName,
      storageNamespacePrefix,
      otherWebView as ExploreUrlWebView,
    );
    sessionStorage.removeItem(key);
  }
}

export function setExploreStateForWebView(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  webView: ExploreUrlWebView,
  state: string,
) {
  sessionStorage.setItem(
    getKeyForSessionStore(exploreName, storageNamespacePrefix, webView),
    state,
  );
}

function shouldSkipSessionStorage(searchParams: URLSearchParams) {
  // exactly one param is set, but it is not `view`
  const hasSingleNonViewParam =
    searchParams.size === 1 && !searchParams.has(ExploreStateURLParams.WebView);

  // exactly 2 params are set and both `view` and `measure` are not set.
  const hasTwoParamsWithoutViewOrMeasure =
    searchParams.size === 2 &&
    !searchParams.has(ExploreStateURLParams.WebView) &&
    !searchParams.has(ExploreStateURLParams.ExpandedMeasure);

  // more than 2 params are set.
  const hasMoreThanTwoParams = searchParams.size > 2;

  return (
    hasSingleNonViewParam ||
    hasTwoParamsWithoutViewOrMeasure ||
    hasMoreThanTwoParams
  );
}
