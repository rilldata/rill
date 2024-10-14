import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type { DashboardMutables } from "./types";

export const setPrimaryDimension = (
  { dashboard }: DashboardMutables,

  dimensionName: string | undefined,
) => {
  dashboard.selectedDimensionName = dimensionName;
  if (dimensionName) {
    dashboard.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
  } else {
    dashboard.activePage = DashboardState_ActivePage.DEFAULT;
  }
};

export const toggleDimensionVisibility = (
  { dashboard, persistentDashboardStore }: DashboardMutables,
  allDimensions: string[],
  dimensionName?: string,
) => {
  if (dimensionName) {
    const deleted = dashboard.visibleDimensionKeys.delete(dimensionName);
    if (!deleted) {
      dashboard.visibleDimensionKeys.add(dimensionName);
    }
  } else {
    const allSelected =
      dashboard.visibleDimensionKeys.size === allDimensions.length;

    dashboard.visibleDimensionKeys = new Set(
      allSelected ? allDimensions.slice(0, 1) : allDimensions,
    );
  }

  dashboard.allDimensionsVisible =
    dashboard.visibleDimensionKeys.size === allDimensions.length;

  persistentDashboardStore.updateVisibleDimensions(
    Array.from(dashboard.visibleDimensionKeys),
  );
};

export const dimensionActions = {
  /**
   * Sets the primary dimension for the dashboard, which
   * activates the dimension table. Setting the primary dimension
   * to undefined closes the dimension table.
   */
  setPrimaryDimension,
  toggleDimensionVisibility,
};
