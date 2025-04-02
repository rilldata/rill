import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import {
  ExploreUrlWebView,
  FromActivePageMap,
  FromURLParamViewMap,
  ToURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import {
  type V1ExploreSpec,
  V1ExploreWebView,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { convertExploreStateToURLSearchParams } from "web-common/src/features/dashboards/url-state/convertExploreStateToURLSearchParams";
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
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  timeControlsState: TimeControlState | undefined,
) {
  const apiWebView =
    FromActivePageMap[exploreState.activePage] ??
    V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE;
  const curWebView = ToURLParamViewMap[apiWebView];
  // type guard
  if (!curWebView) return;

  // Build the url search params for the entire state
  const urlSearchParams = convertExploreStateToURLSearchParams(
    exploreState,
    exploreSpec,
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
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
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
        metricsViewSpec,
        exploreSpec,
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
