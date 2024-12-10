import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import {
  contextColWidthDefaults,
  type MetricsExplorerEntity,
} from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getPersistentDashboardStateForKey } from "@rilldata/web-common/features/dashboards/stores/persistent-dashboard-state";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";

// TODO: Remove this in favour of just `getBasePreset`
export function getDefaultExploreState(
  name: string,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
  defaultExplorePreset = getDefaultExplorePreset(explore, fullTimeRange),
): MetricsExplorerEntity {
  const { partialExploreState } = convertPresetToExploreState(
    metricsView,
    explore,
    defaultExplorePreset,
  );
  return {
    // fields filled here are the ones that are not stored and loaded to/from URL
    name,
    dimensionFilterExcludeMode: new Map(),
    leaderboardContextColumn: LeaderboardContextColumn.HIDDEN,

    temporaryFilterName: null,
    contextColumnWidths: { ...contextColWidthDefaults },

    lastDefinedScrubRange: undefined,

    ...partialExploreState,
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
