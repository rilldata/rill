import type { DashboardMutables } from "./types";

export const toggleMeasureVisibility = (
  { dashboard }: DashboardMutables,
  allMeasures: string[],
  measureName?: string,
) => {
  if (measureName) {
    const deleted = dashboard.visibleMeasureKeys.delete(measureName);
    if (!deleted) {
      dashboard.visibleMeasureKeys.add(measureName);
    } else if (
      dashboard.leaderboardMeasureName === measureName &&
      dashboard.visibleMeasureKeys.size > 0
    ) {
      dashboard.leaderboardMeasureName = dashboard.visibleMeasureKeys
        .keys()
        .next().value;
    }
  } else {
    const allSelected =
      dashboard.visibleMeasureKeys.size === allMeasures.length;

    dashboard.visibleMeasureKeys = new Set(
      allSelected ? allMeasures.slice(0, 1) : allMeasures,
    );
  }

  dashboard.allMeasuresVisible =
    dashboard.visibleMeasureKeys.size === allMeasures.length;
};

export const measureActions = {
  toggleMeasureVisibility,
};
