import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { DashboardDataSources } from "./types";

export const sortingSelectors = {
  /**
   * Gets the sort type for the dash (value, percent, delta, etc.)
   */
  sortType: ({ dashboard }: DashboardDataSources) =>
    dashboard.dashboardSortType,

  /**
   * true if the dashboard is sorted ascending, false otherwise.
   */
  sortedAscending: ({ dashboard }: DashboardDataSources) =>
    dashboard.sortDirection === SortDirection.ASCENDING,

  /**
   * Returns the measure name that the dashboard is sorted by,
   * or null if the dashboard is sorted by dimension value.
   */
  sortMeasure: ({ dashboard }: DashboardDataSources) =>
    dashboard.dashboardSortType !== SortType.DIMENSION &&
    dashboard.dashboardSortType !== SortType.UNSPECIFIED
      ? dashboard.leaderboardMeasureNames[0]
      : null,

  /**
   * Returns true if the dashboard is sorted by a dimension, false otherwise.
   */
  sortedByDimensionValue: ({ dashboard }: DashboardDataSources) =>
    dashboard.dashboardSortType === SortType.DIMENSION,
};
