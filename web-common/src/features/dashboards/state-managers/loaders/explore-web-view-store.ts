import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import {
  copyUrlSearchParamsForView,
  type ExploreUrlWebView,
} from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-specific-url-params";
import {
  FromActivePageMap,
  FromURLParamViewMap,
  ToURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import {
  type V1ExploreSpec,
  V1ExploreWebView,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

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
export function updateExploreSessionStorage(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  explore: V1ExploreSpec,
  timeControlsState: TimeControlState | undefined,
  exploreState: MetricsExplorerEntity,
) {
  const apiWebView =
    FromActivePageMap[exploreState.activePage] ??
    V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE;
  const curWebView = ToURLParamViewMap[apiWebView] as
    | ExploreUrlWebView
    | undefined;
  // type guard
  if (!curWebView) return;

  // Build the url search params for the entire state
  const urlSearchParams = convertExploreStateToURLSearchParams(
    exploreState,
    explore,
    timeControlsState,
    {},
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

export function getPartialExploreStateForWebView(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  urlWebView: ExploreUrlWebView,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
) {
  const key = getKeyForSessionStore(
    exploreName,
    storageNamespacePrefix,
    urlWebView,
  );

  try {
    const storedUrlSearch = sessionStorage.getItem(key);
    if (!storedUrlSearch) return undefined;
    const storedUrlSearchParams = new URLSearchParams(storedUrlSearch);

    const { partialExploreState: storedExploreState } =
      convertURLSearchParamsToExploreState(
        storedUrlSearchParams,
        metricsView,
        explore,
        {},
      );

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
