import type { DashboardMutables } from "./types";
import { getPersistentDashboardStore } from "../../stores/persistent-dashboard-state";

const persistentDashboardStore = getPersistentDashboardStore();

export const toggleMeasureVisibility = (
  { dashboard }: DashboardMutables,

  measureName: string,
) => {
  const deleted = dashboard.visibleMeasureKeys.delete(measureName);

  if (!deleted) {
    dashboard.visibleMeasureKeys.add(measureName);
  }

  persistentDashboardStore.updateVisibleMeasures(
    Array.from(dashboard.visibleMeasureKeys),
  );
};

export const setVisibleMeasures = (
  { dashboard }: DashboardMutables,
  measureNames: string[],
) => {
  dashboard.visibleMeasureKeys = new Set(measureNames);

  persistentDashboardStore.updateVisibleMeasures(
    Array.from(dashboard.visibleMeasureKeys),
  );
};

export const measureActions = {
  setVisibleMeasures,
  toggleMeasureVisibility,
};
