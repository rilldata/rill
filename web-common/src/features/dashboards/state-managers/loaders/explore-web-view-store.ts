import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { toEphemeralMeasuresParam } from "@rilldata/web-common/features/dashboards/ephemeral-measures/url-param";
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
 * Returns the full params of the current view.
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
  if (!curWebView) return undefined;

  // Build the url search params for the entire state
  const urlSearchParams = convertPartialExploreStateToUrlParams(
    exploreSpec,
    metricsViewSpec,
    exploreState,
    timeControlsState,
  );
  // The address bar only carries the ad-hoc measures the current page shows.
  // The session store keeps every definition, since a stored view's own params
  // (visible measures, pivot chips, ...) are validated against the definitions
  // in the same entry when the view is restored; a narrowed list would drop
  // the measures another view still shows.
  if (exploreState.ephemeralMeasures?.length) {
    urlSearchParams.set(
      ExploreStateURLParams.EphemeralMeasures,
      toEphemeralMeasuresParam(exploreState.ephemeralMeasures),
    );
  }
  // Store the full url for the web view
  setExploreStateForWebView(
    exploreName,
    storageNamespacePrefix,
    curWebView,
    urlSearchParams.toString(),
  );

  // We need to make sure other views are in sync
  for (const otherWebViewStr in FromURLParamViewMap) {
    const otherWebView = otherWebViewStr as ExploreUrlWebView;
    if (otherWebView === curWebView) continue;

    const otherWebViewKey = getKeyForSessionStore(
      exploreName,
      storageNamespacePrefix,
      otherWebView,
    );
    let otherWebViewRawSearch = "";
    try {
      otherWebViewRawSearch = sessionStorage.getItem(otherWebViewKey) ?? "";
    } catch {
      // Reading storage throws when the browser blocks it, e.g. in an embed with third-party cookies blocked.
    }
    const otherWebViewUrlParams = new URLSearchParams(otherWebViewRawSearch);

    // Copy relevant params from current view to the other view
    copyUrlSearchParamsForView(
      curWebView,
      urlSearchParams,
      otherWebView,
      otherWebViewUrlParams,
    );

    // Store the full url for the other web view
    setExploreStateForWebView(
      exploreName,
      storageNamespacePrefix,
      otherWebView,
      otherWebViewUrlParams.toString(),
    );
  }

  return urlSearchParams;
}

/**
 * Returns the params stored when the user last visited the view in the url.
 * 1. If no param is set then load the params for the default view from session storage.
 * 2. If only view param is set then load the params from session storage.
 * 3. If view=ttd and `measure` is the only other param set load from session storage.
 */
export function getExploreUrlParamsFromSessionStorage(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  searchParams: URLSearchParams,
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

  let storedUrlSearch: string | null = null;
  try {
    storedUrlSearch = sessionStorage.getItem(key);
  } catch {
    // Reading storage throws when the browser blocks it, e.g. in an embed with third-party cookies blocked.
  }
  if (!storedUrlSearch) return undefined;
  const storedUrlParams = new URLSearchParams(storedUrlSearch);

  // TDD is different from other views. It has a variable that is expanded measure.
  // So we need to copy over the actual measure from current url but keep other params.
  if (viewFromUrl === ExploreUrlWebView.TimeDimension) {
    storedUrlParams.set(
      ExploreStateURLParams.ExpandedMeasure,
      searchParams.get(ExploreStateURLParams.ExpandedMeasure) ?? "",
    );
  }

  return storedUrlParams;
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
  try {
    sessionStorage.setItem(
      getKeyForSessionStore(exploreName, storageNamespacePrefix, webView),
      state,
    );
  } catch {
    // Writing storage throws when the quota is exceeded or the browser blocks it, e.g. in an embed.
  }
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
