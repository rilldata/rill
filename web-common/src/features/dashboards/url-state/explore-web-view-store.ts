import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertExploreStateToPreset } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToPreset";
import { FromActivePageMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
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
    "timeGrain",
    "comparisonDimension",
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
    "timeGrain",
    "comparisonDimension",
  ],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT]: [
    "view",
    "pivotCols",
    "pivotRows",
    "pivotSortAsc",
    "pivotSortBy",
    "pivotRowJoinType",
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
// Keys shared between views.
const ExploreViewSharedKeys = {} as Record<
  V1ExploreWebView,
  Record<V1ExploreWebView, (keyof V1ExplorePreset)[]>
>;
// keys shared between views but to be ignored because they are set exclusively
const ExploreViewIgnoredKeysForShared: (keyof V1ExplorePreset)[] = ["view"];
for (const webView in ExploreViewOtherKeys) {
  const keys = new Set(ExploreViewKeys[webView]);
  ExploreViewSharedKeys[webView] = {};
  const otherKeys = new Set<keyof V1ExplorePreset>();

  for (const otherWebView in ExploreViewKeys) {
    if (webView === otherWebView) continue;
    ExploreViewSharedKeys[webView][otherWebView] = [];

    for (const key of ExploreViewKeys[otherWebView]) {
      if (keys.has(key)) {
        if (!ExploreViewIgnoredKeysForShared.includes(key)) {
          ExploreViewSharedKeys[webView][otherWebView].push(key);
        }
        continue;
      }
      otherKeys.add(key);
    }
  }
  ExploreViewOtherKeys[webView] = [...otherKeys];
}
// Values shared across the views. Any keys not defined in ExploreViewKeys will fall under this.
// Having a catch-all like this will avoid issues where new fields added are not lost.
const SharedStateStoreKey = "__shared";

export function getKeyForSessionStore(
  exploreName: string,
  prefix: string | undefined,
  view: string,
) {
  return `rill:app:explore:${prefix ?? ""}${exploreName}:${view}`.toLowerCase();
}

export function updateExploreSessionStore(
  exploreName: string,
  prefix: string | undefined,
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  timeControlsState: TimeControlState | undefined,
) {
  const view = FromActivePageMap[exploreState.activePage];
  const key = getKeyForSessionStore(exploreName, prefix, view);
  const sharedKey = getKeyForSessionStore(
    exploreName,
    prefix,
    SharedStateStoreKey,
  );

  const preset = convertExploreStateToPreset(
    exploreState,
    exploreSpec,
    timeControlsState,
  );
  const storedPreset: V1ExplorePreset = {};
  const sharedPreset: V1ExplorePreset = {
    ...preset,
  };

  for (const key of ExploreViewKeys[view]) {
    storedPreset[key] = preset[key] as any;
    delete sharedPreset[key];
  }
  for (const key of ExploreViewOtherKeys[view]) {
    storedPreset[key] = preset[key] as any;
    delete sharedPreset[key];
  }

  sessionStorage.setItem(key, JSON.stringify(storedPreset));
  sessionStorage.setItem(sharedKey, JSON.stringify(sharedPreset));

  for (const otherView in ExploreViewSharedKeys[view]) {
    const sharedKeys = ExploreViewSharedKeys[view][otherView];
    if (!sharedKeys?.length) continue;

    const otherViewKey = getKeyForSessionStore(exploreName, prefix, otherView);
    const otherViewRawPreset = sessionStorage.getItem(otherViewKey) ?? "{}";

    try {
      const otherViewPreset = JSON.parse(otherViewRawPreset) as V1ExplorePreset;
      for (const sharedKey of sharedKeys) {
        if (!(sharedKey in storedPreset)) continue;
        otherViewPreset[sharedKey] = storedPreset[sharedKey];
      }
      sessionStorage.setItem(otherViewKey, JSON.stringify(otherViewPreset));
    } catch {
      // ignore errors
    }
  }
}

export function clearExploreSessionStore(
  exploreName: string,
  prefix: string | undefined,
) {
  for (const view in ExploreViewKeys) {
    const key = getKeyForSessionStore(exploreName, prefix, view);
    sessionStorage.removeItem(key);
  }

  const sharedKey = getKeyForSessionStore(
    exploreName,
    prefix,
    SharedStateStoreKey,
  );
  sessionStorage.removeItem(sharedKey);
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
    const sharedPreset = JSON.parse(sharedRawPreset) as V1ExplorePreset;
    const parsedPreset = JSON.parse(rawPreset) as V1ExplorePreset;
    return {
      view,
      ...sharedPreset,
      ...parsedPreset,
    };
  } catch {
    return undefined;
  }
}

export function hasSessionStorageData(
  exploreName: string,
  prefix: string | undefined,
) {
  const sharedKey = getKeyForSessionStore(
    exploreName,
    prefix,
    SharedStateStoreKey,
  );
  return !!sessionStorage.getItem(sharedKey);
}
