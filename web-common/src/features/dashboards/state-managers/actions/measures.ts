import type { DashboardMutables } from "./types";

export const toggleMeasureVisibility = (
  { dashboard }: DashboardMutables,
  allMeasures: string[],
  measureName: string,
) => {
  const index = dashboard.visibleMeasures.indexOf(measureName);
  if (index > -1) {
    dashboard.visibleMeasures.splice(index, 1);
    if (
      dashboard.leaderboardSortByMeasureName === measureName &&
      dashboard.visibleMeasures.length > 1
    ) {
      dashboard.leaderboardSortByMeasureName = dashboard.visibleMeasures[0];
    }
  } else {
    dashboard.visibleMeasures.push(measureName);
  }

  dashboard.allMeasuresVisible =
    dashboard.visibleMeasures.length === allMeasures.length;
};

export const toggleAllMeasuresVisibility = (
  { dashboard }: DashboardMutables,
  allMeasures: string[],
) => {
  const allSelected = dashboard.visibleMeasures.length === allMeasures.length;

  dashboard.visibleMeasures = allSelected
    ? allMeasures.slice(0, 1)
    : [...allMeasures];
  dashboard.allMeasuresVisible = !dashboard.allMeasuresVisible;
};

export const setMeasureVisibility = (
  { dashboard }: DashboardMutables,
  measures: string[],
  allMeasures: string[],
) => {
  dashboard.visibleMeasures = measures;

  // If the current leaderboard measure is hidden, select a new one from visible measures
  if (
    !measures.includes(dashboard.leaderboardSortByMeasureName) &&
    measures.length > 0
  ) {
    dashboard.leaderboardSortByMeasureName = measures[0];
  }

  // Update leaderboardMeasureNames to only include visible measures
  if (dashboard.leaderboardMeasureNames) {
    dashboard.leaderboardMeasureNames =
      dashboard.leaderboardMeasureNames.filter((name) =>
        measures.includes(name),
      );
    // If no leaderboard measures are visible, set to the first visible measure
    if (dashboard.leaderboardMeasureNames.length === 0 && measures.length > 0) {
      dashboard.leaderboardMeasureNames = [measures[0]];
    }
  }

  dashboard.allMeasuresVisible =
    dashboard.visibleMeasures.length === allMeasures.length;
};

export const measureActions = {
  toggleMeasureVisibility,
  toggleAllMeasuresVisibility,
  setMeasureVisibility,
};
