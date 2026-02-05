import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";

/**
 * In-memory map: exploreName → partial ExploreState (JSON string).
 *
 * Replaces localStorage-based persistence for "most recent explore state."
 * Within the same browser session, navigating away from and back to an explore
 * will restore the last-visited state from this map.
 * On page refresh, the in-memory state is lost and YAML defaults apply.
 *
 * Only a curated subset of JSON-safe fields is saved — specifically sort/visibility
 * preferences. Filters, time ranges, and other stateful fields are NOT saved here;
 * they come from URL params, session storage, or YAML defaults.
 */
const lastVisitedState = new Map<string, string>();

export function getLastVisitedState(exploreName: string): string | undefined {
  return lastVisitedState.get(exploreName);
}

/**
 * Saves a curated subset of the explore state to in-memory storage.
 * Only saves sort/visibility preferences (all JSON-safe primitives).
 * Does NOT save filters, time ranges, or other complex objects (Set, Map, etc.).
 */
export function setLastVisitedState(
  exploreName: string,
  exploreState: ExploreState,
) {
  if (
    exploreState.activePage !== DashboardState_ActivePage.DEFAULT &&
    exploreState.activePage !== DashboardState_ActivePage.DIMENSION_TABLE
  ) {
    // We are not saving any state for non-explore pages
    return;
  }

  const subset: Partial<ExploreState> = {
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
  };

  lastVisitedState.set(exploreName, JSON.stringify(subset));
}

export function clearLastVisitedState(exploreName: string) {
  lastVisitedState.delete(exploreName);
}

/** For testing only: directly set the raw JSON string in the map */
export function setLastVisitedStateRaw(exploreName: string, rawJson: string) {
  lastVisitedState.set(exploreName, rawJson);
}
