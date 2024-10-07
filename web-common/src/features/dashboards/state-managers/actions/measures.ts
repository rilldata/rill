import type { DashboardMutables } from "./types";

export const toggleMeasureVisibility = (
  { dashboard }: DashboardMutables,

  measureName: string,
) => {
  const deleted = dashboard.visibleMeasureKeys.delete(measureName);

  if (!deleted) {
    dashboard.visibleMeasureKeys.add(measureName);
  }
};

export const setVisibleMeasures = (
  { dashboard }: DashboardMutables,
  measureNames: string[],
) => {
  dashboard.visibleMeasureKeys = new Set(measureNames);
};

export const measureActions = {
  setVisibleMeasures,
  toggleMeasureVisibility,
};
