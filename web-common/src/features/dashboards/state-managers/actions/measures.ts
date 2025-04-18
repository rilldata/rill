import type { DashboardMutables } from "./types";

export const toggleMeasureVisibility = (
  { dashboard }: DashboardMutables,
  allMeasures: string[],
  measureName: string,
) => {
  const index = dashboard.visibleMeasures.indexOf(measureName);
  if (index > -1) {
    dashboard.visibleMeasures.splice(index, 1);
    // If we're hiding the current sort measure or it's the only visible measure, select a new one
    if (
      (dashboard.leaderboardSortByMeasureName === measureName ||
        dashboard.visibleMeasures.length === 0) &&
      dashboard.visibleMeasures.length > 0
    ) {
      // Find the next visible measure in leaderboardMeasureNames order
      const nextMeasure = dashboard.leaderboardMeasureNames?.find(
        (name) =>
          name !== measureName && dashboard.visibleMeasures.includes(name),
      );
      if (nextMeasure) {
        dashboard.leaderboardSortByMeasureName = nextMeasure;
      } else {
        // Fallback to first visible measure if no leaderboard measure is available
        dashboard.leaderboardSortByMeasureName = dashboard.visibleMeasures[0];
      }
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
    // Find the next visible measure in leaderboardMeasureNames order
    const nextMeasure = dashboard.leaderboardMeasureNames?.find((name) =>
      measures.includes(name),
    );
    if (nextMeasure) {
      dashboard.leaderboardSortByMeasureName = nextMeasure;
    } else {
      // Fallback to first visible measure if no leaderboard measure is available
      dashboard.leaderboardSortByMeasureName = measures[0];
    }
  }

  // Update leaderboard measure names to only include visible measures, maintaining their order
  if (dashboard.leaderboardMeasureNames) {
    dashboard.leaderboardMeasureNames = dashboard.leaderboardMeasureNames
      .filter((name) => measures.includes(name))
      .sort((a, b) => measures.indexOf(a) - measures.indexOf(b));
  }

  dashboard.allMeasuresVisible =
    dashboard.visibleMeasures.length === allMeasures.length;
};

export const measureActions = {
  toggleMeasureVisibility,
  toggleAllMeasuresVisibility,
  setMeasureVisibility,
};
