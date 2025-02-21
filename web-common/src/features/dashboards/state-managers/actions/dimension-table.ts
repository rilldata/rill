import { SortType } from "../../proto-state/derived-types";
import { toggleSort, sortActions } from "./sorting";
import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { setContextColumn } from "./context-columns";
import { setLeaderboardMeasureNames } from "./core-actions";
import type { DashboardMutables } from "./types";

export const handleMeasureColumnHeaderClick = (
  generalArgs: DashboardMutables,
  measureName: string,
) => {
  const { leaderboardMeasureNames: names } = generalArgs.dashboard;

  if (measureName === names[0] + "_delta") {
    toggleSort(generalArgs, SortType.DELTA_ABSOLUTE);
    setContextColumn(generalArgs, LeaderboardContextColumn.DELTA_ABSOLUTE);
  } else if (measureName === names[0] + "_delta_perc") {
    toggleSort(generalArgs, SortType.DELTA_PERCENT);
    setContextColumn(generalArgs, LeaderboardContextColumn.DELTA_PERCENT);
  } else if (measureName === names[0] + "_percent_of_total") {
    toggleSort(generalArgs, SortType.PERCENT);
    setContextColumn(generalArgs, LeaderboardContextColumn.PERCENT);
  } else if (measureName === names[0]) {
    toggleSort(generalArgs, SortType.VALUE);
  } else {
    setLeaderboardMeasureNames(generalArgs, [measureName]);
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
  handleMeasureColumnHeaderClick,
};
