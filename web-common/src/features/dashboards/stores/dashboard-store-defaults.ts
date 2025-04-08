import {
  contextColWidthDefaults,
  LeaderboardContextColumn,
} from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { type MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getDefaultTimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils";
import {
  getDefaultTimeRange,
  getDefaultTimeZone,
} from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { ToURLParamTimeGrainMapMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import type {
  V1ExplorePreset,
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";

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
    whereFilter: createAndExpression([]),
    dimensionThresholdFilters: [],
    dimensionsWithInlistFilter: [],

    temporaryFilterName: null,
    contextColumnWidths: { ...contextColWidthDefaults },

    lastDefinedScrubRange: undefined,

    ...partialInitState,
  } as MetricsExplorerEntity;
}
