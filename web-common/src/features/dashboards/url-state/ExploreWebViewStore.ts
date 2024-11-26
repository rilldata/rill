import { page } from "$app/stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertMetricsEntityToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertMetricsEntityToURLSearchParams";
import { convertMetricsExploreToPreset } from "@rilldata/web-common/features/dashboards/url-state/convertMetricsExploreToPreset";
import { convertPresetToMetricsExplore } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToMetricsExplore";
import { FromActivePageMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { sessionStorageStore } from "@rilldata/web-common/lib/store-utils/session-storage";
import { mergeSearchParams } from "@rilldata/web-common/lib/url-utils";
import {
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export const ExploreWebViewNonPivot = "NON_PIVOT";
type ExploreWebView = V1ExploreWebView | typeof ExploreWebViewNonPivot;
const ExploreViewKeys: Record<ExploreWebView, (keyof V1ExplorePreset)[]> = {
  [V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED]: [],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW]: [
    "overviewExpandedDimension",
    "overviewSortBy",
    "overviewSortAsc",
    "overviewSortType",
  ],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION]: [
    "timeDimensionMeasure",
    "timeDimensionChartType",
    "timeDimensionPin",
  ],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT]: [
    "pivotCols",
    "pivotRows",
    "pivotSortAsc",
    "pivotSortBy",
  ],
  [ExploreWebViewNonPivot]: [],
  [V1ExploreWebView.EXPLORE_WEB_VIEW_CANVAS]: [],
};
ExploreViewKeys[ExploreWebViewNonPivot] = [
  ...ExploreViewKeys[V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW],
  ...ExploreViewKeys[V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION],
  ...ExploreViewKeys[V1ExploreWebView.EXPLORE_WEB_VIEW_CANVAS],
];
// keys other than the current web view
const ExploreViewOtherKeys: Record<ExploreWebView, (keyof V1ExplorePreset)[]> =
  {
    [V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED]: [],
    [V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW]: [],
    [V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION]: [],
    [V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT]: [],
    [ExploreWebViewNonPivot]: [],
    [V1ExploreWebView.EXPLORE_WEB_VIEW_CANVAS]: [],
  };
for (const webView in ExploreViewOtherKeys) {
  const keys = new Set(ExploreViewKeys[webView]);
  const otherKeys: (keyof V1ExplorePreset)[] = [];
  for (const otherWebView in ExploreViewKeys) {
    for (const key of ExploreViewKeys[otherWebView]) {
      if (keys.has(key)) continue;
      otherKeys.push(key);
    }
  }
  ExploreViewOtherKeys[webView] = otherKeys;
}

function getStoreForExploreWebView(exploreName: string, view: ExploreWebView) {
  const key = `rill:app:explore:${exploreName}:${view}`;
  return sessionStorageStore<V1ExplorePreset>(key, {});
}

export class ExploreWebViewStore {
  public readonly stores: Record<
    ExploreWebView,
    ReturnType<typeof getStoreForExploreWebView>
  >;

  public constructor(exploreName: string) {
    this.stores = {
      [V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED]:
        getStoreForExploreWebView(
          exploreName,
          V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED,
        ),
      [V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW]: getStoreForExploreWebView(
        exploreName,
        V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW,
      ),
      [V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION]:
        getStoreForExploreWebView(
          exploreName,
          V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
        ),
      [V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT]: getStoreForExploreWebView(
        exploreName,
        V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
      ),
      [ExploreWebViewNonPivot]: getStoreForExploreWebView(
        exploreName,
        ExploreWebViewNonPivot,
      ),
      [V1ExploreWebView.EXPLORE_WEB_VIEW_CANVAS]: getStoreForExploreWebView(
        exploreName,
        V1ExploreWebView.EXPLORE_WEB_VIEW_CANVAS,
      ),
    };
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
      );
    }
    this.updateStoreForView(view, exploreState, exploreSpec);
  }

  public getUrlForView(
    view: ExploreWebView,
    exploreState: MetricsExplorerEntity,
    metricsSpec: V1MetricsViewSpec,
    exploreSpec: V1ExploreSpec,
    basePreset: V1ExplorePreset,
    additionaPreset: V1ExplorePreset = {},
  ) {
    const currentPreset = convertMetricsExploreToPreset(
      exploreState,
      exploreSpec,
    );
    const preset = {
      ...currentPreset,
      ...get(this.stores[view]),
      ...additionaPreset,
    };

    for (const key of ExploreViewOtherKeys[view]) {
      preset[key] = basePreset[key] as any;
    }
    if (view === ExploreWebViewNonPivot) {
      preset.view = undefined;
    } else {
      preset.view = view;
    }

    const { partialExploreState } = convertPresetToMetricsExplore(
      metricsSpec,
      exploreSpec,
      preset,
    );
    const searchParams = convertMetricsEntityToURLSearchParams(
      partialExploreState as MetricsExplorerEntity,
      exploreSpec,
      basePreset,
    );
    const u = new URL(get(page).url);
    // clear any existing search params.
    u.search = "";
    mergeSearchParams(searchParams, u.searchParams);
    return u.pathname + u.search;
  }

  private updateStoreForView(
    view: ExploreWebView,
    exploreState: MetricsExplorerEntity,
    exploreSpec: V1ExploreSpec,
  ) {
    const store = this.stores[view];
    const preset = convertMetricsExploreToPreset(exploreState, exploreSpec);
    const storedPreset: V1ExplorePreset = {};

    for (const key of ExploreViewKeys[view]) {
      storedPreset[key] = preset[key] as any;
    }

    store.set(storedPreset);
  }
}
