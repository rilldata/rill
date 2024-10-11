import type { DashboardMutables } from "./types";

export const toggleMeasureVisibility = (
  { dashboard, persistentDashboardStore }: DashboardMutables,
  allMeasures: string[],
  measureName?: string,
) => {
  if (measureName) {
    const deleted = dashboard.visibleMeasureKeys.delete(measureName);
    if (!deleted) {
      dashboard.visibleMeasureKeys.add(measureName);
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

  persistentDashboardStore.updateVisibleMeasures(
    Array.from(dashboard.visibleMeasureKeys),
  );
};

export const measureActions = {
  toggleMeasureVisibility,
};
