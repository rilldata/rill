import { validateAndCleanExploreState } from "@rilldata/web-common/features/dashboards/stores/validate-and-clean-explore-state";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import type { ExploreState } from "web-common/src/features/dashboards/stores/explore-state";

function getKeyForLocalStore(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
) {
  return `rill:app:explore:${storageNamespacePrefix ?? ""}${exploreName}`.toLowerCase();
}

export function getMostRecentPartialExploreState(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
) {
  const key = getKeyForLocalStore(exploreName, storageNamespacePrefix);
  try {
    const rawExploreState = localStorage.getItem(key);
    if (!rawExploreState) {
      // Return this so that destructuring is simple
      return { mostRecentPartialExploreState: undefined, errors: [] };
    }

    const stateFromLocalStorage = JSON.parse(
      rawExploreState,
    ) as Partial<ExploreState>;
    const errors = validateAndCleanExploreState(
      metricsViewSpec,
      exploreSpec,
      stateFromLocalStorage,
    );
    return { mostRecentPartialExploreState: stateFromLocalStorage, errors };
  } catch {
    // no-op
  }
  // Return this so that destructuring is simple
  return { mostRecentPartialExploreState: undefined, errors: [] };
}

/**
 * We only want to save a subset of the most recent explore state in the long term local storage.
 * This method selects the exact fields we want to save in the local storage.
 */
export function saveMostRecentPartialExploreState(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  exploreState: ExploreState,
) {
  if (
    exploreState.activePage !== DashboardState_ActivePage.DEFAULT &&
    exploreState.activePage !== DashboardState_ActivePage.DIMENSION_TABLE
  ) {
    // We are not saving any state for non-explore pages
    return;
  }

  try {
    setMostRecentExploreStateInLocalStorage(
      exploreName,
      storageNamespacePrefix,
      {
        selectedTimezone: exploreState.selectedTimezone,

        visibleMeasures: exploreState.visibleMeasures,
        allMeasuresVisible: exploreState.allMeasuresVisible,
        visibleDimensions: exploreState.visibleDimensions,
        allDimensionsVisible: exploreState.allDimensionsVisible,

        leaderboardSortByMeasureName: exploreState.leaderboardSortByMeasureName,
        leaderboardMeasureNames: exploreState.leaderboardMeasureNames,
        leaderboardShowContextForAllMeasures:
          exploreState.leaderboardShowContextForAllMeasures,
        sortDirection: exploreState.sortDirection,
        dashboardSortType: exploreState.dashboardSortType,
      },
    );
  } catch {
    // no-op
  }
}

export function clearMostRecentExploreState(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
) {
  const key = getKeyForLocalStore(exploreName, storageNamespacePrefix);
  localStorage.removeItem(key);
}

export function setMostRecentExploreStateInLocalStorage(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  exploreState: Partial<ExploreState>,
) {
  localStorage.setItem(
    getKeyForLocalStore(exploreName, storageNamespacePrefix),
    JSON.stringify(exploreState),
  );
}
