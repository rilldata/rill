import { browser } from "$app/environment";
import { page } from "$app/stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertExploreStateToPreset } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToPreset";
import {
  FromActivePageMap,
  ToURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { sessionStorageStore } from "@rilldata/web-common/lib/store-utils/session-storage";
import {
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export const ExploreWebViewNonPivot = "NON_PIVOT";
type ExploreWebView = V1ExploreWebView | typeof ExploreWebViewNonPivot;
const ExploreViewKeys: Record<ExploreWebView, (keyof V1ExplorePreset)[]> = {
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
  [ExploreWebViewNonPivot]: [],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_CANVAS]: [],
};
ExploreViewKeys[ExploreWebViewNonPivot] = [
  ...ExploreViewKeys[V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE],
  ...ExploreViewKeys[V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION],
  ...ExploreViewKeys[V1ExploreWebView.EXPLORE_WEB_VIEW_CANVAS],
];
// keys other than the current web view
const ExploreViewOtherKeys: Record<ExploreWebView, (keyof V1ExplorePreset)[]> =
  {
    [V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED]: [],
    [V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE]: [],
    [V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION]: [],
    [V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT]: [],
    [ExploreWebViewNonPivot]: [],
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

function getKeyForSessionStore(
  exploreName: string,
  prefix: string | undefined,
  view: string,
) {
  return `rill:app:explore:${prefix ?? ""}${exploreName}:${view}`;
}

export class ExploreWebViewStore {
  public readonly stores: Record<
    ExploreWebView,
    ReturnType<typeof sessionStorageStore<V1ExplorePreset>>
  >;

  public constructor(exploreName: string, prefix: string | undefined) {
    this.stores = {
      [V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED]:
        sessionStorageStore<V1ExplorePreset>(
          getKeyForSessionStore(
            exploreName,
            prefix,
            ToURLParamViewMap[V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED]!,
          ),
          {},
        ),
      [V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE]:
        sessionStorageStore<V1ExplorePreset>(
          getKeyForSessionStore(
            exploreName,
            prefix,
            ToURLParamViewMap[V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE]!,
          ),
          {},
        ),
      [V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION]:
        sessionStorageStore<V1ExplorePreset>(
          getKeyForSessionStore(
            exploreName,
            prefix,
            ToURLParamViewMap[
              V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION
            ]!,
          ),
          {},
        ),
      [V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT]:
        sessionStorageStore<V1ExplorePreset>(
          getKeyForSessionStore(
            exploreName,
            prefix,
            ToURLParamViewMap[V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT]!,
          ),
          {},
        ),
      [ExploreWebViewNonPivot]: sessionStorageStore<V1ExplorePreset>(
        getKeyForSessionStore(
          exploreName,
          prefix,
          ToURLParamViewMap[ExploreWebViewNonPivot],
        ),
        {},
      ),
      [V1ExploreWebView.EXPLORE_WEB_VIEW_CANVAS]:
        sessionStorageStore<V1ExplorePreset>(
          getKeyForSessionStore(
            exploreName,
            prefix,
            ToURLParamViewMap[V1ExploreWebView.EXPLORE_WEB_VIEW_CANVAS]!,
          ),
          {},
        ),
    };
  }

  public static hasPresetForView(
    exploreName: string,
    prefix: string | undefined,
    view: string,
  ) {
    if (!browser) return false;
    const key = getKeyForSessionStore(exploreName, prefix, view);
    return !!sessionStorage.getItem(key);
  }

  public static getPresetForView(
    exploreName: string,
    prefix: string | undefined,
    view: string,
  ): V1ExplorePreset {
    if (!browser) return {};
    const key = getKeyForSessionStore(exploreName, prefix, view);
    const stored = sessionStorage.getItem(key);
    if (!stored) return {};
    try {
      const parsed = JSON.parse(stored);
      return parsed as V1ExplorePreset;
    } catch {
      // ignore
    }
    return {};
  }

  public updateStores(
    exploreState: MetricsExplorerEntity,
    exploreSpec: V1ExploreSpec,
  ) {
    const view = FromActivePageMap[exploreState.activePage];
    if (view !== V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT) {
      this.updateStoreForView(
        ExploreWebViewNonPivot,
        exploreState,
        exploreSpec,
        view,
      );
    }
    this.updateStoreForView(view, exploreState, exploreSpec);
  }

  public static getUrlForView(pageUrl: URL, view: V1ExploreWebView) {
    const u = new URL(pageUrl);
    u.search = "";
    u.searchParams.set(ExploreStateURLParams.WebView, ToURLParamViewMap[view]!);
    return u.pathname + u.search;
  }

  private updateStoreForView(
    storeView: ExploreWebView,
    exploreState: MetricsExplorerEntity,
    exploreSpec: V1ExploreSpec,
    forView = storeView,
  ) {
    const store = this.stores[storeView];
    const preset = convertExploreStateToPreset(exploreState, exploreSpec);
    const storedPreset: V1ExplorePreset = {
      ...preset,
    };

    for (const key of ExploreViewOtherKeys[forView]) {
      delete storedPreset[key];
    }

    store.set(storedPreset);
  }
}
