import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { MetricsExplorerEntity } from "../../dashboard-stores";

const toggleSort =
  (sortType: SortType) => (metricsExplorer: MetricsExplorerEntity) => {
    // if sortType is not provided,  or if it is provided
    // and is the same as the current sort type,
    // then just toggle the current sort direction
    if (
      sortType === undefined ||
      metricsExplorer.dashboardSortType === sortType
    ) {
      metricsExplorer.sortDirection =
        metricsExplorer.sortDirection === SortDirection.ASCENDING
          ? SortDirection.DESCENDING
          : SortDirection.ASCENDING;
    } else {
      // if the sortType is different from the current sort type,
      //  then update the sort type and set the sort direction
      // to descending
      metricsExplorer.dashboardSortType = sortType;
      metricsExplorer.sortDirection = SortDirection.DESCENDING;
    }
  };

export const sortActions = {
  toggleSort,
  sortByDimensionValue: () => toggleSort(SortType.DIMENSION),
  setSortDescending: () => (metricsExplorer) => {
    metricsExplorer.sortDirection = SortDirection.DESCENDING;
  },
};
