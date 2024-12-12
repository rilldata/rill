import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import {
  contextColWidthDefaults,
  type MetricsExplorerEntity,
} from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getPersistentDashboardStateForKey } from "@rilldata/web-common/features/dashboards/stores/persistent-dashboard-state";
import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";

// TODO: Remove this in favour of just `getBasePreset`
export function getDefaultExploreState(
  name: string,
  initState: Partial<MetricsExplorerEntity>,
): MetricsExplorerEntity {
  return {
    // fields filled here are the ones that are not stored and loaded to/from URL
    name,
    dimensionFilterExcludeMode: new Map(),
    leaderboardContextColumn: LeaderboardContextColumn.HIDDEN,

    temporaryFilterName: null,
    contextColumnWidths: { ...contextColWidthDefaults },

    lastDefinedScrubRange: undefined,

    ...initState,
  } as MetricsExplorerEntity;
}

export function restorePersistedDashboardState(
  exploreSpec: V1ExploreSpec,
  key: string,
) {
  const stateFromLocalStorage = getPersistentDashboardStateForKey(key);
  if (!stateFromLocalStorage) return undefined;

  const partialExploreState: Partial<MetricsExplorerEntity> = {};

  if (stateFromLocalStorage.visibleMeasures) {
    partialExploreState.allMeasuresVisible =
      stateFromLocalStorage.visibleMeasures.length ===
      exploreSpec.measures?.length;
    partialExploreState.visibleMeasureKeys = new Set(
      stateFromLocalStorage.visibleMeasures,
    );
  }
  if (stateFromLocalStorage.visibleDimensions) {
    partialExploreState.allDimensionsVisible =
      stateFromLocalStorage.visibleDimensions.length ===
      exploreSpec.dimensions?.length;
    partialExploreState.visibleDimensionKeys = new Set(
      stateFromLocalStorage.visibleDimensions,
    );
  }
  if (stateFromLocalStorage.leaderboardMeasureName) {
    partialExploreState.leaderboardMeasureName =
      stateFromLocalStorage.leaderboardMeasureName;
  }
  if (stateFromLocalStorage.dashboardSortType) {
    partialExploreState.dashboardSortType =
      stateFromLocalStorage.dashboardSortType;
  }
  if (stateFromLocalStorage.sortDirection) {
    partialExploreState.sortDirection = stateFromLocalStorage.sortDirection;
  }

  return partialExploreState;
}
