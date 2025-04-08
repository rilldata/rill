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

export const leaderboardMeasureCount = ({
  dashboard,
}: DashboardDataSources) => {
  return dashboard.leaderboardMeasureCount ?? 1;
};

export const leaderboardShowAllMeasures = ({
  dashboard,
}: DashboardDataSources) => {
  return dashboard.leaderboardShowAllMeasures ?? false;
};

// @deprecated
export const activeMeasuresFromMeasureCount = (
  dashboardDataSources: DashboardDataSources,
): string[] => {
  const { validMetricsView, validExplore, dashboard } = dashboardDataSources;
  if (!validMetricsView?.measures || !validExplore?.measures) return [];

  const visibleMeasureSpecs = visibleMeasures(dashboardDataSources);

  return visibleMeasureSpecs
    .slice(0, dashboard.leaderboardMeasureCount ?? 1)
    .map(({ name }) => name)
    .filter((name): name is string => name !== undefined);
};

export const leaderboardSelectors = {
  leaderboardSortByMeasureName,

  leaderboardMeasureNames,

  leaderboardMeasureCount,

  leaderboardShowAllMeasures,

  activeMeasuresFromMeasureCount,
};
