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
      dashboard.leaderboardMeasureName === measureName &&
      dashboard.visibleMeasures.length > 1
    ) {
      dashboard.leaderboardMeasureName = dashboard.visibleMeasures[0];
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

export const measureActions = {
  toggleMeasureVisibility,
  toggleAllMeasuresVisibility,
};
