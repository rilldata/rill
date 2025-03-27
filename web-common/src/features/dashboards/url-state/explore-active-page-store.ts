import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertExploreStateToURLSearchParamsNoCompression } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
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
  ],
  [DashboardState_ActivePage.DIMENSION_TABLE]: [],
  [DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL]: [
    "activePage",
    "selectedComparisonDimension",
    "tdd",
  ],
  [DashboardState_ActivePage.PIVOT]: ["activePage", "pivot"],
};
ExploreActivePageKeys[DashboardState_ActivePage.DIMENSION_TABLE] =
  ExploreActivePageKeys[DashboardState_ActivePage.DEFAULT];

// keys other than the current web view
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
// Keys shared between views.
const ExploreActivePageSharedKeys = {} as Record<
  DashboardState_ActivePage,
  (keyof MetricsExplorerEntity)[]
>;
// keys shared between views but to be ignored because they are set exclusively
const ExploreViewIgnoredKeysForShared: (keyof MetricsExplorerEntity)[] = [
  "activePage",
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
        if (!ExploreViewIgnoredKeysForShared.includes(key)) {
          ExploreActivePageSharedKeys[activePage][otherActivePage].push(key);
        }
        continue;
      }
      otherKeys.add(key);
    }
  }
  ExploreActivePageOtherKeys[activePage] = [...otherKeys];
}

// Values shared across the views. Any keys not defined in ExploreViewKeys will fall under this.
// Having a catch-all like this will avoid issues where new fields added are not lost.
const SharedStateStoreKey = 0;

export function getKeyForSessionStore(
  exploreName: string,
  prefix: string | undefined,
  activePage: number,
) {
  return `rill:app:explore:${prefix ?? ""}${exploreName}:${activePage}`.toLowerCase();
}

// TODO: revisit namings
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

  const keyForActivePage = getKeyForSessionStore(
    exploreName,
    prefix,
    activePage,
  );
  const sharedKey = getKeyForSessionStore(
    exploreName,
    prefix,
    SharedStateStoreKey,
  );

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

  const storedUrlSearch = convertExploreStateToURLSearchParamsNoCompression(
    storedExploreState as MetricsExplorerEntity,
    explore,
    timeControlsState,
    {},
  );
  const sharedUrlSearch = convertExploreStateToURLSearchParamsNoCompression(
    sharedExploreState as MetricsExplorerEntity,
    explore,
    timeControlsState,
    {},
  );
  try {
    sessionStorage.setItem(keyForActivePage, storedUrlSearch.toString());
    sessionStorage.setItem(sharedKey, sharedUrlSearch.toString());
  } catch {
    // no-op
  }

  for (const otherView in ExploreActivePageSharedKeys[activePage]) {
    const sharedKeys = ExploreActivePageSharedKeys[activePage][otherView];
    if (!sharedKeys?.length) continue;

    const otherViewKey = getKeyForSessionStore(
      exploreName,
      prefix,
      Number(otherView),
    );
    const otherViewRawPreset = sessionStorage.getItem(otherViewKey) ?? "";

    const { partialExploreState: otherViewExploreState } =
      convertURLSearchParamsToExploreState(
        new URLSearchParams(otherViewRawPreset),
        metricsView,
        explore,
        {},
      );

    for (const sharedKey of sharedKeys) {
      if (!(sharedKey in storedExploreState)) continue;
      otherViewExploreState[sharedKey] = storedExploreState[sharedKey];
    }
    const otherViewUrlSearch =
      convertExploreStateToURLSearchParamsNoCompression(
        otherViewExploreState as MetricsExplorerEntity,
        explore,
        timeControlsState,
        {},
      );
    try {
      sessionStorage.setItem(otherViewKey, otherViewUrlSearch.toString());
    } catch {
      // no-op
    }
  }
}

export function clearExploreSessionStore(
  exploreName: string,
  prefix: string | undefined,
) {
  for (const view in ExploreActivePageKeys) {
    const key = getKeyForSessionStore(exploreName, prefix, Number(view));
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
    const rawUrlSearch = sessionStorage.getItem(key) ?? "";
    const { partialExploreState: sharedExploreState } =
      convertURLSearchParamsToExploreState(
        new URLSearchParams(sharedUrlSearch),
        metricsView,
        explore,
        {},
      );
    const { partialExploreState: parsedExploreState } =
      convertURLSearchParamsToExploreState(
        new URLSearchParams(rawUrlSearch),
        metricsView,
        explore,
        {},
      );

    if (
      activePage === DashboardState_ActivePage.DEFAULT &&
      parsedExploreState.selectedDimensionName
    ) {
      activePage = DashboardState_ActivePage.DIMENSION_TABLE;
    }

    return <Partial<MetricsExplorerEntity>>{
      activePage,
      ...sharedExploreState,
      ...parsedExploreState,
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
