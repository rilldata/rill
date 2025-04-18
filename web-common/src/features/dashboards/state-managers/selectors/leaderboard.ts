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

  // Filter and sort the leaderboard measure names based on the order in visibleMeasures
  const filteredNames = dashboard.leaderboardMeasureNames
    ?.filter((name) => visibleMeasureNames.has(name))
    .sort((a, b) => {
      const aIndex = visibleMeasuresList.findIndex((m) => m.name === a);
      const bIndex = visibleMeasuresList.findIndex((m) => m.name === b);
      return aIndex - bIndex;
    });

  return filteredNames?.length
    ? filteredNames
    : [dashboard.leaderboardSortByMeasureName];
};

export const leaderboardShowContextForAllMeasures = ({
  dashboard,
}: DashboardDataSources) => {
  return dashboard.leaderboardShowContextForAllMeasures ?? false;
};

export const leaderboardSelectors = {
  leaderboardSortByMeasureName,
  leaderboardMeasureNames,
  leaderboardShowContextForAllMeasures,
};
