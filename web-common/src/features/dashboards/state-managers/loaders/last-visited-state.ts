import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";

/**
 * In-memory map: exploreName → partial ExploreState snapshot.
 *
 * Replaces localStorage-based persistence for "most recent explore state."
 * Within the same browser session, navigating away from and back to an explore
 * will restore the last-visited state from this map.
 * On page refresh, the in-memory state is lost and YAML defaults apply.
 */
const lastVisitedState = new Map<string, Partial<ExploreState>>();

export function getLastVisitedState(
  exploreName: string,
): Partial<ExploreState> | undefined {
  return lastVisitedState.get(exploreName);
}

/**
 * Saves a snapshot of the explore state to in-memory storage.
 * Copies the fields we care about so that later mutations to the live
 * ExploreState don't affect the stored snapshot.
 */
export function setLastVisitedState(
  exploreName: string,
  exploreState: ExploreState,
) {
  // Only snapshot state from the main explore or dimension table views.
  // TDD and pivot views are handled by the per-view state store
  // (explore-web-view-store.ts), so we skip them here to avoid
  // overwriting the user's explore-view preferences with TDD/pivot state.
  if (
    exploreState.activePage !== DashboardState_ActivePage.DEFAULT &&
    exploreState.activePage !== DashboardState_ActivePage.DIMENSION_TABLE
  ) {
    return;
  }

  const subset: Partial<ExploreState> = {
    // Filters — deep copy protobuf objects to prevent mutation
    whereFilter: structuredClone(exploreState.whereFilter),
    dimensionThresholdFilters: structuredClone(
      exploreState.dimensionThresholdFilters,
    ),
    dimensionsWithInlistFilter: [...exploreState.dimensionsWithInlistFilter],
    pinnedFilters: new Set(exploreState.pinnedFilters),
    dimensionFilterExcludeMode: new Map(
      exploreState.dimensionFilterExcludeMode,
    ),

    // Time — deep copy to snapshot Date objects and nested fields
    selectedTimeRange: structuredClone(exploreState.selectedTimeRange),
    selectedComparisonTimeRange: structuredClone(
      exploreState.selectedComparisonTimeRange,
    ),
    showTimeComparison: exploreState.showTimeComparison,
    selectedComparisonDimension: exploreState.selectedComparisonDimension,
    selectedTimezone: exploreState.selectedTimezone,

    // Sort / visibility — shallow copy arrays
    visibleMeasures: [...exploreState.visibleMeasures],
    allMeasuresVisible: exploreState.allMeasuresVisible,
    visibleDimensions: [...exploreState.visibleDimensions],
    allDimensionsVisible: exploreState.allDimensionsVisible,
    leaderboardSortByMeasureName: exploreState.leaderboardSortByMeasureName,
    leaderboardMeasureNames: [...exploreState.leaderboardMeasureNames],
    leaderboardShowContextForAllMeasures:
      exploreState.leaderboardShowContextForAllMeasures,
    sortDirection: exploreState.sortDirection,
    dashboardSortType: exploreState.dashboardSortType,

    // View
    activePage: exploreState.activePage,
    selectedDimensionName: exploreState.selectedDimensionName,
  };

  lastVisitedState.set(exploreName, subset);
}

export function clearLastVisitedState(exploreName: string) {
  lastVisitedState.delete(exploreName);
}

/** For testing only: directly set a partial explore state in the map */
export function setLastVisitedStateRaw(
  exploreName: string,
  state: Partial<ExploreState>,
) {
  lastVisitedState.set(exploreName, state);
}
