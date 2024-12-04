import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import {
  contextColWidthDefaults,
  type MetricsExplorerEntity,
} from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getPersistentDashboardState } from "@rilldata/web-common/features/dashboards/stores/persistent-dashboard-state";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { getLocalUserPreferencesState } from "@rilldata/web-common/features/dashboards/user-preferences";
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
  defaultExplorePreset = getDefaultExplorePreset(
    explore,
    getLocalUserPreferencesState(name),
    fullTimeRange,
  ),
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
  metricsExplorer: MetricsExplorerEntity,
) {
  const persistedState = getPersistentDashboardState();
  if (persistedState.visibleMeasures) {
    metricsExplorer.allMeasuresVisible =
      persistedState.visibleMeasures.length ===
      metricsExplorer.visibleMeasureKeys.size; // TODO: check values
    metricsExplorer.visibleMeasureKeys = new Set(
      persistedState.visibleMeasures,
    );
  }
  if (persistedState.visibleDimensions) {
    metricsExplorer.allDimensionsVisible =
      persistedState.visibleDimensions.length ===
      metricsExplorer.visibleDimensionKeys.size; // TODO: check values
    metricsExplorer.visibleDimensionKeys = new Set(
      persistedState.visibleDimensions,
    );
  }
  if (persistedState.leaderboardMeasureName) {
    metricsExplorer.leaderboardMeasureName =
      persistedState.leaderboardMeasureName;
  }
  if (persistedState.dashboardSortType) {
    metricsExplorer.dashboardSortType = persistedState.dashboardSortType;
  }
  if (persistedState.sortDirection) {
    metricsExplorer.sortDirection = persistedState.sortDirection;
  }
  return metricsExplorer;
}
