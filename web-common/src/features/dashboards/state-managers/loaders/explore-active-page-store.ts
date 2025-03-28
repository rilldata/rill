import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { AD_BIDS_EXPLORE_NAME } from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type V1ExploreSpec,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

const ExploreActivePageKeys: Record<
  DashboardState_ActivePage,
  (keyof MetricsExplorerEntity)[]
> = {
  [DashboardState_ActivePage.UNSPECIFIED]: [],
  [DashboardState_ActivePage.DEFAULT]: [
    "activePage",
    "visibleMeasureKeys",
    "allMeasuresVisible",
    "visibleDimensionKeys",
    "allDimensionsVisible",
    "selectedComparisonDimension",
    "selectedDimensionName",
    "leaderboardMeasureName",
    "sortDirection",
    "leaderboardContextColumn",
    "leaderboardMeasureCount",
    "selectedScrubRange",
  ],
  [DashboardState_ActivePage.DIMENSION_TABLE]: [],
  [DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL]: [
    "activePage",
    "selectedComparisonDimension",
    "tdd",
    "selectedScrubRange",
  ],
  [DashboardState_ActivePage.PIVOT]: ["activePage", "pivot"],
};
ExploreActivePageKeys[DashboardState_ActivePage.DIMENSION_TABLE] =
  ExploreActivePageKeys[DashboardState_ActivePage.DEFAULT];

// Keys other than the current active page.
const ExploreActivePageOtherKeys: Record<
  DashboardState_ActivePage,
  (keyof MetricsExplorerEntity)[]
> = {
  [DashboardState_ActivePage.UNSPECIFIED]: [],
  [DashboardState_ActivePage.DEFAULT]: [],
  [DashboardState_ActivePage.DIMENSION_TABLE]: [],
  [DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL]: [],
  [DashboardState_ActivePage.PIVOT]: [],
};
// Keys shared between pages.
const ExploreActivePageSharedKeys = {} as Record<
  DashboardState_ActivePage,
  (keyof MetricsExplorerEntity)[]
>;
// Common time range url params to remove since we have to use timeControlsStore.
// TODO: revisit this when we get rid of timeControlsStore
const ExploreCommonTimeRangeURLParams: ExploreStateURLParams[] = [
  ExploreStateURLParams.TimeRange,
  ExploreStateURLParams.TimeGrain,
  ExploreStateURLParams.ComparisonTimeRange,
  ExploreStateURLParams.TimeZone,
];

// Build ExploreActivePageOtherKeys and ExploreActivePageSharedKeys based on ExploreActivePageKeys
for (const activePage in ExploreActivePageOtherKeys) {
  const keys = new Set(ExploreActivePageKeys[activePage]);
  ExploreActivePageSharedKeys[activePage] = {};
  const otherKeys = new Set<keyof MetricsExplorerEntity>();

  for (const otherActivePage in ExploreActivePageKeys) {
    if (activePage === otherActivePage) continue;
    ExploreActivePageSharedKeys[activePage][otherActivePage] = [];

    for (const key of ExploreActivePageKeys[otherActivePage]) {
      if (keys.has(key)) {
        // activePage is set explicitly
        if (key !== "activePage") {
          ExploreActivePageSharedKeys[activePage][otherActivePage].push(key);
        }
        continue;
      }
      otherKeys.add(key);
    }
  }
  ExploreActivePageOtherKeys[activePage] = [...otherKeys];
}

// Values shared across the pages. Any keys not defined in ExploreActivePageKeys will fall under this.
// Having a catch-all like this will avoid issues where new fields added are not lost.
const SharedStateStoreKey = 0;

export function getKeyForSessionStore(
  exploreName: string,
  prefix: string | undefined,
  activePage: number,
) {
  return `rill:app:explore:${prefix ?? ""}${exploreName}:${activePage}`.toLowerCase();
}

export function updateExploreSessionStore(
  exploreName: string,
  prefix: string | undefined,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  timeControlsState: TimeControlState | undefined,
  exploreState: MetricsExplorerEntity,
) {
  let activePage = exploreState.activePage;
  if (activePage === DashboardState_ActivePage.DIMENSION_TABLE) {
    activePage = DashboardState_ActivePage.DEFAULT;
  }

  const storedExploreState: Partial<MetricsExplorerEntity> = {};
  const sharedExploreState: Partial<MetricsExplorerEntity> = {
    ...exploreState,
  };

  for (const key of ExploreActivePageKeys[activePage]) {
    storedExploreState[key] = exploreState[key] as any;
    delete sharedExploreState[key];
  }
  for (const key of ExploreActivePageOtherKeys[activePage]) {
    storedExploreState[key] = exploreState[key] as any;
    delete sharedExploreState[key];
  }
  storedExploreState.activePage = Number(storedExploreState.activePage);

  const storedUrlSearch = convertExploreStateToURLSearchParams(
    storedExploreState as MetricsExplorerEntity,
    explore,
    timeControlsState,
    {},
  );
  const sharedUrlSearch = convertExploreStateToURLSearchParams(
    sharedExploreState as MetricsExplorerEntity,
    explore,
    timeControlsState,
    {},
  );
  try {
    setExploreStateForActivePage(
      exploreName,
      prefix,
      activePage,
      storedUrlSearch.toString(),
    );
    setSharedExploreState(exploreName, prefix, sharedUrlSearch.toString());
  } catch {
    // no-op
  }

  for (const otherPage in ExploreActivePageSharedKeys[activePage]) {
    const sharedKeys = ExploreActivePageSharedKeys[activePage][otherPage];
    if (!sharedKeys?.length) continue;

    const otherPageKey = getKeyForSessionStore(
      exploreName,
      prefix,
      Number(otherPage),
    );
    const otherPageRawPreset = sessionStorage.getItem(otherPageKey) ?? "";

    const { partialExploreState: otherPageExploreState } =
      convertURLSearchParamsToExploreState(
        new URLSearchParams(otherPageRawPreset),
        metricsView,
        explore,
        {},
      );

    for (const sharedKey of sharedKeys) {
      if (!(sharedKey in storedExploreState)) continue;
      otherPageExploreState[sharedKey] = storedExploreState[sharedKey];
    }
    const otherPageUrlSearch = convertExploreStateToURLSearchParams(
      otherPageExploreState as MetricsExplorerEntity,
      explore,
      timeControlsState,
      {},
    );
    try {
      setExploreStateForActivePage(
        exploreName,
        prefix,
        Number(otherPage),
        otherPageUrlSearch.toString(),
      );
    } catch {
      // no-op
    }
  }
}

export function getExplorePresetForActivePage(
  exploreName: string,
  prefix: string | undefined,
  activePage: DashboardState_ActivePage,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
) {
  const key = getKeyForSessionStore(exploreName, prefix, activePage);
  const sharedKey = getKeyForSessionStore(
    exploreName,
    prefix,
    SharedStateStoreKey,
  );

  try {
    const sharedUrlSearch = sessionStorage.getItem(sharedKey);
    if (!sharedUrlSearch) return undefined;
    const storedUrlSearch = sessionStorage.getItem(key) ?? "";

    const sharedUrlSearchParams = new URLSearchParams(sharedUrlSearch);
    const storedUrlSearchParams = new URLSearchParams(storedUrlSearch);

    ExploreCommonTimeRangeURLParams.forEach((p) =>
      storedUrlSearchParams.delete(p),
    );

    const { partialExploreState: sharedExploreState } =
      convertURLSearchParamsToExploreState(
        sharedUrlSearchParams,
        metricsView,
        explore,
        {},
      );
    const { partialExploreState: storedExploreState } =
      convertURLSearchParamsToExploreState(
        storedUrlSearchParams,
        metricsView,
        explore,
        {},
      );

    if (
      activePage === DashboardState_ActivePage.DEFAULT &&
      storedExploreState.selectedDimensionName
    ) {
      activePage = DashboardState_ActivePage.DIMENSION_TABLE;
    }

    return <Partial<MetricsExplorerEntity>>{
      activePage,
      ...sharedExploreState,
      ...storedExploreState,
    };
  } catch {
    return undefined;
  }
}

export function clearExploreSessionStore(
  exploreName: string,
  prefix: string | undefined,
) {
  for (const activePage in ExploreActivePageKeys) {
    const key = getKeyForSessionStore(exploreName, prefix, Number(activePage));
    sessionStorage.removeItem(key);
  }

  const sharedKey = getKeyForSessionStore(
    exploreName,
    prefix,
    SharedStateStoreKey,
  );
  sessionStorage.removeItem(sharedKey);
}

export function setExploreStateForActivePage(
  exploreName: string,
  prefix: string | undefined,
  activePage: DashboardState_ActivePage,
  state: string,
) {
  sessionStorage.setItem(
    getKeyForSessionStore(exploreName, prefix, activePage),
    state,
  );
}

export function setSharedExploreState(
  exploreName: string,
  prefix: string | undefined,
  state: string,
) {
  sessionStorage.setItem(
    getKeyForSessionStore(exploreName, prefix, SharedStateStoreKey),
    state,
  );
}
