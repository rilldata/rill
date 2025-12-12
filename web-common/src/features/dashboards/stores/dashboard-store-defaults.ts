import {
  contextColWidthDefaults,
  LeaderboardContextColumn,
} from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import { type ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";

// TODO: Remove this in favour of just `getDefaultExplorePreset`
export function getFullInitExploreState(
  name: string,
  partialInitState: Partial<ExploreState>,
): ExploreState {
  return {
    // fields filled here are the ones that are not stored and loaded to/from URL
    name,
    dimensionFilterExcludeMode: new Map(),
    leaderboardContextColumn: LeaderboardContextColumn.HIDDEN,
    pinnedFilters: new Set<string>(),
    temporaryFilterName: null,
    contextColumnWidths: { ...contextColWidthDefaults },

    lastDefinedScrubRange: undefined,

    ...partialInitState,
  } as ExploreState;
}
