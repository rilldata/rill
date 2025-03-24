import { SortType } from "../../proto-state/derived-types";
import { toggleSort, sortActions } from "./sorting";
import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { setContextColumn } from "./context-columns";
import type { DashboardMutables } from "./types";
import { setLeaderboardMeasureName } from "./core-actions";

// export const handleDimensionMeasureColumnHeaderClick = (
//   generalArgs: DashboardMutables,
//   measureName: string,
// ) => {
//   console.log("handleDimensionMeasureColumnHeaderClick: ", measureName);

//   const { dashboard } = generalArgs;

//   if (measureName === dashboard.sortedMeasureName + "_delta") {
//     toggleSort(generalArgs, SortType.DELTA_ABSOLUTE);
//     setContextColumn(generalArgs, LeaderboardContextColumn.DELTA_ABSOLUTE);
//   } else if (measureName === dashboard.sortedMeasureName + "_delta_perc") {
//     toggleSort(generalArgs, SortType.DELTA_PERCENT);
//     setContextColumn(generalArgs, LeaderboardContextColumn.DELTA_PERCENT);
//   } else if (
//     measureName ===
//     dashboard.sortedMeasureName + "_percent_of_total"
//   ) {
//     toggleSort(generalArgs, SortType.PERCENT);
//     setContextColumn(generalArgs, LeaderboardContextColumn.PERCENT);
//   } else if (measureName === dashboard.sortedMeasureName) {
//     toggleSort(generalArgs, SortType.VALUE);
//   } else {
//     // If clicking on a different measure, update the sorted measure and sort by it
//     dashboard.sortedMeasureName = measureName;
//     toggleSort(generalArgs, SortType.VALUE);
//     sortActions.setSortDescending(generalArgs);
//   }
// };

export const handleDimensionMeasureColumnHeaderClick = (
  generalArgs: DashboardMutables,
  measureName: string,
) => {
  const { leaderboardMeasureName: name } = generalArgs.dashboard;

  if (measureName === name + "_delta") {
    toggleSort(generalArgs, SortType.DELTA_ABSOLUTE, name);
    setContextColumn(generalArgs, LeaderboardContextColumn.DELTA_ABSOLUTE);
  } else if (measureName === name + "_delta_perc") {
    toggleSort(generalArgs, SortType.DELTA_PERCENT, name);
    setContextColumn(generalArgs, LeaderboardContextColumn.DELTA_PERCENT);
  } else if (measureName === name + "_percent_of_total") {
    toggleSort(generalArgs, SortType.PERCENT, name);
    setContextColumn(generalArgs, LeaderboardContextColumn.PERCENT);
  } else if (measureName === name) {
    toggleSort(generalArgs, SortType.VALUE);
  } else {
    setLeaderboardMeasureName(generalArgs, measureName);
    toggleSort(generalArgs, SortType.VALUE);
    sortActions.setSortDescending(generalArgs);
  }
};

export const dimensionTableActions = {
  /**
   * handles clicking on a measure column header in the dimension
   * table, including the delta, delta percent, and percent of total
   * columns. This will set the active measure and sort the leaderboard
   * by the selected measure (or context column).
   */
  handleDimensionMeasureColumnHeaderClick,
};
