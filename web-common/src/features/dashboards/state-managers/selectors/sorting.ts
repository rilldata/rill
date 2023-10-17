import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { SelectorFnArgs } from "./types";

export const sortingSelectors = {
  /**
   * Gets the sort type for the dash (value, percent, delta, etc.)
   */
  sortType: ([dashboard, _]: SelectorFnArgs) => dashboard.dashboardSortType,

  /**
   * true if the dashboard is sorted ascending, false otherwise.
   */
  sortedAscending: ([dashboard, _]: SelectorFnArgs) =>
    dashboard.sortDirection === SortDirection.ASCENDING,

  /**
   * Returns the measure name that the dashboard is sorted by,
   * or null if the dashboard is sorted by dimension value.
   */
  sortMeasure: ([dashboard, _]: SelectorFnArgs) =>
    dashboard.dashboardSortType !== SortType.DIMENSION &&
    dashboard.dashboardSortType !== SortType.UNSPECIFIED
      ? dashboard.leaderboardMeasureName
      : null,

  /**
   * Returns true if the dashboard is sorted by a dimension, false otherwise.
   */
  sortedByDimensionValue: ([dashboard, _]: SelectorFnArgs) =>
    dashboard.dashboardSortType === SortType.DIMENSION,
};
