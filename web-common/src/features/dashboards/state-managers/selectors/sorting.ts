import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";

export const sortingSelectors = {
  /**
   * Gets the sort type for the dash (value, percent, delta, etc.)
   */
  sortType: (dashboard: MetricsExplorerEntity) => dashboard.dashboardSortType,

  sortedAscending: (dashboard: MetricsExplorerEntity) =>
    dashboard.sortDirection === SortDirection.ASCENDING,

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
