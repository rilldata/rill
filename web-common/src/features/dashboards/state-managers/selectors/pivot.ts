import type { DashboardDataSources } from "./types";

export const pivotSelectors = {
  showPivot: ({ dashboard }: DashboardDataSources) => dashboard.pivot.active,
  rows: ({ dashboard }: DashboardDataSources) => dashboard.pivot.rows,
  columns: ({ dashboard }: DashboardDataSources) => dashboard.pivot.columns,
};
