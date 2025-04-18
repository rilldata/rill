import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { SortType } from "../../proto-state/derived-types";
import { setContextColumn } from "./context-columns";
import { setLeaderboardSortByMeasureName } from "./leaderboard";
import { sortActions, toggleSort } from "./sorting";
import type { DashboardMutables } from "./types";

export const handleDimensionMeasureColumnHeaderClick = (
  generalArgs: DashboardMutables,
  measureName: string,
) => {
  const { leaderboardSortByMeasureName: name } = generalArgs.dashboard;

  const delta = name + "_delta";
  const deltaPerc = name + "_delta_perc";
  const percentOfTotal = name + "_percent_of_total";

  switch (measureName) {
    case delta:
      toggleSort(generalArgs, SortType.DELTA_ABSOLUTE, name);
      setContextColumn(generalArgs, LeaderboardContextColumn.DELTA_ABSOLUTE);
      break;
    case deltaPerc:
      toggleSort(generalArgs, SortType.DELTA_PERCENT, name);
      setContextColumn(generalArgs, LeaderboardContextColumn.DELTA_PERCENT);
      break;
    case percentOfTotal:
      toggleSort(generalArgs, SortType.PERCENT, name);
      setContextColumn(generalArgs, LeaderboardContextColumn.PERCENT);
      break;
    case name:
      toggleSort(generalArgs, SortType.VALUE);
      break;
    default:
      setLeaderboardSortByMeasureName(generalArgs, measureName);
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
