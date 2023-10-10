import { SortDirection, SortType } from "../../proto-state/derived-types";
import type { MetricsExplorerEntity } from "../../dashboard-stores";
import { toggleSort } from "./sorting";
import { LeaderboardContextColumn } from "../../leaderboard-context-column";
import { setContextColumn } from "./context-columns";

export const handleMeasureColumnHeaderClick =
  (measureName: string) => (dash: MetricsExplorerEntity) => {
    const { leaderboardMeasureName: name } = dash;

    if (measureName === name + "_delta") {
      toggleSort(SortType.DELTA_ABSOLUTE)(dash);
      setContextColumn(LeaderboardContextColumn.DELTA_ABSOLUTE)(dash);
    } else if (measureName === name + "_delta_perc") {
      toggleSort(SortType.DELTA_PERCENT)(dash);
      setContextColumn(LeaderboardContextColumn.DELTA_PERCENT)(dash);
    } else if (measureName === name + "_percent_of_total") {
      toggleSort(SortType.PERCENT)(dash);
      setContextColumn(LeaderboardContextColumn.PERCENT)(dash);
    } else if (measureName === name) {
      toggleSort(SortType.VALUE)(dash);
    } else {
      setLeaderboardMeasureName(measureName);
      toggleSort(SortType.VALUE)(dash);
      setSortDescending();
    }
  };

export const dimTableActions = {
  toggleSort,
  sortByDimensionValue: () => toggleSort(SortType.DIMENSION),
  setSortDescending: () => (metricsExplorer) => {
    metricsExplorer.sortDirection = SortDirection.DESCENDING;
  },
};
