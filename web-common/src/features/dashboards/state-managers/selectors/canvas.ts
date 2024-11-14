import type { DashboardDataSources } from "./types";

export const canvasSelectors = {
  showCanvas: ({ dashboard }: DashboardDataSources) => dashboard.canvas.active,
};
