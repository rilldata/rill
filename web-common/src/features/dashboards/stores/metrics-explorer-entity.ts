import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import type { PivotState } from "@rilldata/web-common/features/dashboards/pivot/types";
import type {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { TDDState } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import type {
  DashboardTimeControls,
  ScrubRange,
} from "@rilldata/web-common/lib/time/types";
import type { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

export interface DimensionThresholdFilter {
  name: string;
  filter: V1Expression;
}

export interface MetricsExplorerEntity {
  name: string;

  /**
   * This array controls which measures are visible in
   * explorer on the client. Note that this will need to be
   * updated to include all measure keys upon initialization
   * or else all measure will be hidden
   */
  visibleMeasureKeys: Set<string>;

  /**
   * While the `visibleMeasureKeys` has the list of visible measures,
   * this is explicitly needed to fill the state.
   * TODO: clean this up when we refactor how url state is synced
   */
  allMeasuresVisible: boolean;

  /**
   * This array controls which dimensions are visible in
   * explorer on the client.Note that if this is null, all
   * dimensions will be visible (this is needed to default to all visible
   * when there are not existing keys in the URL or saved on the
   * server)
   */
  visibleDimensionKeys: Set<string>;

  /**
   * While the `visibleDimensionKeys` has the list of all visible dimensions,
   * this is explicitly needed to fill the state.
   * TODO: clean this up when we refactor how url state is synced
   */
  allDimensionsVisible: boolean;

  /**
   * This is the name of the primary active measure in the dashboard.
   * This is the measure that will be shown in leaderboards, and
   * will be used for sorting the leaderboard and dimension detail table.
   * This "name" is the internal name of the measure from the YAML,
   * not the human readable name.
   */
  leaderboardMeasureName: string;

  /**
   * This is the sort type that will be used for the leaderboard
   * and dimension detail table. See SortType for more details.
   */
  dashboardSortType: SortType;

  /**
   * This is the sort direction that will be used for the leaderboard
   * and dimension detail table.
   */
  sortDirection: SortDirection;

  whereFilter: V1Expression;
  havingFilter: V1Expression;
  dimensionThresholdFilters: Array<DimensionThresholdFilter>;

  /**
   * stores whether a dimension is in include/exclude filter mode
   * false/absence = include, true = exclude
   */
  dimensionFilterExcludeMode: Map<string, boolean>;

  /**
   * Used to add a dropdown for newly added dimension/measure filters.
   * Such filter will not have an entry in where/having expression objects.
   */
  temporaryFilterName: string | null;

  /**
   * user selected time range
   */
  selectedTimeRange?: DashboardTimeControls;

  /**
   * user selected scrub range
   */
  selectedScrubRange?: ScrubRange;
  lastDefinedScrubRange?: ScrubRange;

  selectedComparisonTimeRange?: DashboardTimeControls;
  selectedComparisonDimension?: string;

  /**
   * Explicit active page setting.
   * This avoids using presence of value in `selectedDimensionName` or `expandedMeasureName`.
   */
  activePage: DashboardState_ActivePage;

  /**
   * user selected timezone, should default to "UTC" if no other value is set
   */
  selectedTimezone: string;

  /**
   * Search text state for dimension tables. This search text state
   * is shared by both the dimension detail table AND the time
   * detailed dimension table, so that the same filter will be
   * applied when switching between those views.
   */
  dimensionSearchText?: string;

  /**
   * flag to show/hide time comparison based on user preference.
   * This controls whether a time comparison is shown in e.g.
   * the line charts and bignums.
   * It does NOT affect the leaderboard context column.
   */
  showTimeComparison?: boolean;

  /**
   * state of context column in the leaderboard
   */
  leaderboardContextColumn: LeaderboardContextColumn;

  /**
   * Width of each context column. Needs to be reset to default
   * when changing context column or switching between leaderboard
   * and dimension detail table
   */
  contextColumnWidths: ContextColWidths;

  /**
   * The name of the dimension that is currently shown in the dimension
   * detail table. If this is undefined, then the dimension detail table
   * is not shown.
   */
  selectedDimensionName?: string;

  /**
   * Consolidated state for Time Dimenstion Detail view
   */
  tdd: TDDState;

  pivot: PivotState;

  proto?: string;
}

export type ContextColWidths = {
  [LeaderboardContextColumn.DELTA_ABSOLUTE]: number;
  [LeaderboardContextColumn.DELTA_PERCENT]: number;
  [LeaderboardContextColumn.PERCENT]: number;
};

export const contextColWidthDefaults: ContextColWidths = {
  [LeaderboardContextColumn.DELTA_ABSOLUTE]: 56,
  [LeaderboardContextColumn.DELTA_PERCENT]: 44,
  [LeaderboardContextColumn.PERCENT]: 44,
};
