import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";

export const sortingSelectors = {
  /**
   * Gets the sort type for the dash (value, percent, delta, etc.)
   */
  sortType: (dashboard: MetricsExplorerEntity) => dashboard.dashboardSortType,

  /**
   * true if the dashboard is sorted ascending, false otherwise.
   */
  sortedAscending: (dashboard: MetricsExplorerEntity) =>
    dashboard.sortDirection === SortDirection.ASCENDING,

  /**
   * Returns the measure name that the dashboard is sorted by,
   * or null if the dashboard is sorted by dimension value.
   */
  sortMeasure: (dashboard: MetricsExplorerEntity) =>
    dashboard.dashboardSortType !== SortType.DIMENSION &&
    dashboard.dashboardSortType !== SortType.UNSPECIFIED
      ? dashboard.leaderboardMeasureName
      : null,

  /**
   * Returns true if the dashboard is sorted by a dimension, false otherwise.
   */
  sortedByDimensionValue: (dashboard: MetricsExplorerEntity) =>
    dashboard.dashboardSortType === SortType.DIMENSION,
};
