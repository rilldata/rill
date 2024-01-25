import type { DashboardDataSources } from "./types";

export const pivotSelectors = {
  showPivot: ({ dashboard }: DashboardDataSources) => dashboard.pivot.active,
};
