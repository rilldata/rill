import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import {
  contextColWidthDefaults,
  type MetricsExplorerEntity,
} from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";

// TODO: Remove this in favour of just `getDefaultExplorePreset`
export function getFullInitExploreState(
  name: string,
  partialInitState: Partial<MetricsExplorerEntity>,
): MetricsExplorerEntity {
  return {
    // fields filled here are the ones that are not stored and loaded to/from URL
    name,
    dimensionFilterExcludeMode: new Map(),
    leaderboardContextColumn: LeaderboardContextColumn.HIDDEN,

    temporaryFilterName: null,
    contextColumnWidths: { ...contextColWidthDefaults },

    lastDefinedScrubRange: undefined,

    ...partialInitState,
  } as MetricsExplorerEntity;
}
