import { SortType } from "../../proto-state/derived-types";
import { toggleSort, sortActions } from "./sorting";
import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { setContextColumn } from "./context-columns";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";
import { setLeaderboardMeasureName } from "./core-actions";

export const handleMeasureColumnHeaderClick = (
  dash: MetricsExplorerEntity,
  measureName: string
) => {
  const { leaderboardMeasureName: name } = dash;

  if (measureName === name + "_delta") {
    toggleSort(dash, SortType.DELTA_ABSOLUTE);
    setContextColumn(dash, LeaderboardContextColumn.DELTA_ABSOLUTE);
  } else if (measureName === name + "_delta_perc") {
    toggleSort(dash, SortType.DELTA_PERCENT);
    setContextColumn(dash, LeaderboardContextColumn.DELTA_PERCENT);
  } else if (measureName === name + "_percent_of_total") {
    toggleSort(dash, SortType.PERCENT);
    setContextColumn(dash, LeaderboardContextColumn.PERCENT);
  } else if (measureName === name) {
    toggleSort(dash, SortType.VALUE);
  } else {
    setLeaderboardMeasureName(dash, measureName);
    toggleSort(dash, SortType.VALUE);
    sortActions.setSortDescending(dash);
  }
};

export const dimTableActions = {
  handleMeasureColumnHeaderClick,
};
