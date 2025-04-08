import { visibleMeasures } from "./measures";
import type { DashboardDataSources } from "./types";

export const leaderboardSortByMeasureName = ({
  dashboard,
}: DashboardDataSources) => {
  return dashboard.leaderboardSortByMeasureName;
};

export const leaderboardMeasureNames = ({
  dashboard,
  ...rest
}: DashboardDataSources) => {
  const visibleMeasuresList = visibleMeasures({ dashboard, ...rest });
  const visibleMeasureNames = new Set(
    visibleMeasuresList.map((measure) => measure.name),
  );

  const filteredNames = dashboard.leaderboardMeasureNames?.filter((name) =>
    visibleMeasureNames.has(name),
  );

  return filteredNames?.length
    ? filteredNames
    : [dashboard.leaderboardSortByMeasureName];
};

export const leaderboardShowAllMeasures = ({
  dashboard,
}: DashboardDataSources) => {
  return dashboard.leaderboardShowAllMeasures ?? false;
};

export const leaderboardSelectors = {
  leaderboardSortByMeasureName,

  leaderboardMeasureNames,

  leaderboardShowAllMeasures,
};
