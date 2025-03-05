import type { DashboardMutables } from "./types";

export const toggleMeasureVisibility = (
  { dashboard }: DashboardMutables,
  allMeasures: string[],
  measureName?: string,
) => {
  if (measureName) {
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
  } else {
    const allSelected = dashboard.visibleMeasures.length === allMeasures.length;

    dashboard.visibleMeasures = allSelected
      ? allMeasures.slice(0, 1)
      : [...allMeasures];
  }

  dashboard.allMeasuresVisible =
    dashboard.visibleMeasures.length === allMeasures.length;
};

export const setMeasureVisibility = (
  { dashboard }: DashboardMutables,
  measures?: string[],
  allMeasures?: string[],
) => {
  dashboard.visibleMeasureKeys = new Set(measures);

  dashboard.allMeasuresVisible =
    dashboard.visibleMeasureKeys.size === allMeasures?.length;
};

export const measureActions = {
  toggleMeasureVisibility,
  setMeasureVisibility,
};
