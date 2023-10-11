import type { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import type {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type {
  DashboardTimeControls,
  ScrubRange,
} from "@rilldata/web-common/lib/time/types";
import type { V1MetricsViewFilter } from "@rilldata/web-common/runtime-client";

export interface MetricsExplorerEntity {
  name: string;
  /**
   * selected measure names to be shown
   */
  selectedMeasureNames: Array<string>;

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

  filters: V1MetricsViewFilter;

  /**
   * stores whether a dimension is in include/exclude filter mode
   * false/absence = include, true = exclude
   */
  dimensionFilterExcludeMode: Map<string, boolean>;

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
   * user selected timezone
   */
  selectedTimezone?: string;

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
   * user selected dimension
   */
  selectedDimensionName?: string;

  proto?: string;

  /**
   * proto for the default set of selections
   */
  defaultProto?: string;
}
