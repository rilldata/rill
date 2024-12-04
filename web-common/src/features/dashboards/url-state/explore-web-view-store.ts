import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertExploreStateToPreset } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToPreset";
import {
  FromActivePageMap,
  ToURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import {
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
} from "@rilldata/web-common/runtime-client";

const ExploreViewKeys: Record<V1ExploreWebView, (keyof V1ExplorePreset)[]> = {
  [V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED]: [],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE]: [
    "view",
    "measures",
    "dimensions",
    "exploreExpandedDimension",
    "exploreSortBy",
    "exploreSortAsc",
    "exploreSortType",
  ],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION]: [
    "view",
    "timeDimensionMeasure",
    "timeDimensionChartType",
    "timeDimensionPin",
  ],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT]: [
    "view",
    "pivotCols",
    "pivotRows",
    "pivotSortAsc",
    "pivotSortBy",
  ],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_CANVAS]: [],
};
// keys other than the current web view
const ExploreViewOtherKeys: Record<
  V1ExploreWebView,
  (keyof V1ExplorePreset)[]
> = {
  [V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED]: [],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE]: [],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION]: [],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT]: [],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_CANVAS]: [],
};
for (const webView in ExploreViewOtherKeys) {
  const keys = new Set(ExploreViewKeys[webView]);
  const otherKeys = new Set<keyof V1ExplorePreset>();
  for (const otherWebView in ExploreViewKeys) {
    for (const key of ExploreViewKeys[otherWebView]) {
      if (keys.has(key)) continue;
      otherKeys.add(key);
    }
  }
  ExploreViewOtherKeys[webView] = [...otherKeys];
}
export const SharedStateStoreKey = "__common";

export function getKeyForSessionStore(
  exploreName: string,
  prefix: string | undefined,
  view: string,
) {
  return `rill:app:explore:${prefix ?? ""}${exploreName}:${view}`;
}

export function updateExploreSessionStore(
  exploreName: string,
  prefix: string | undefined,
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
) {
  const view = FromActivePageMap[exploreState.activePage];
  const key = getKeyForSessionStore(exploreName, prefix, view);
  const sharedKey = getKeyForSessionStore(
    exploreName,
    prefix,
    SharedStateStoreKey,
  );

  const preset = convertExploreStateToPreset(exploreState, exploreSpec);
  const storedPreset: V1ExplorePreset = {};
  const sharedPreset: V1ExplorePreset = {
    ...preset,
  };

  for (const key of ExploreViewKeys[view]) {
    storedPreset[key] = preset[key] as any;
    delete sharedPreset[key];
  }
  for (const key of ExploreViewOtherKeys[view]) {
    delete sharedPreset[key];
  }

  sessionStorage.setItem(key, JSON.stringify(storedPreset));
  sessionStorage.setItem(sharedKey, JSON.stringify(sharedPreset));
}

export function getExplorePresetForWebView(
  exploreName: string,
  prefix: string | undefined,
  view: V1ExploreWebView,
) {
  const key = getKeyForSessionStore(exploreName, prefix, view);
  const sharedKey = getKeyForSessionStore(
    exploreName,
    prefix,
    SharedStateStoreKey,
  );

  const sharedRawPreset = sessionStorage.getItem(sharedKey);
  if (!sharedRawPreset) return undefined;
  const rawPreset = sessionStorage.getItem(key) ?? "{}";
  try {
    const parsedPreset = JSON.parse(rawPreset) as V1ExplorePreset;
    const sharedPreset = JSON.parse(sharedRawPreset) as V1ExplorePreset;
    return {
      view,
      ...sharedPreset,
      ...parsedPreset,
    };
  } catch {
    return undefined;
  }
}

export function getUrlForWebView(
  pageUrl: URL,
  view: V1ExploreWebView,
  extraParams: Record<string, string> = {},
) {
  const u = new URL(pageUrl);
  u.search = "";
  u.searchParams.set(ExploreStateURLParams.WebView, ToURLParamViewMap[view]!);
  for (const param in extraParams) {
    u.searchParams.set(param, extraParams[param]);
  }
  return u.pathname + u.search;
}
