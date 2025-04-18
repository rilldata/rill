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
  { dashboard }: DashboardMutables,
  allDimensions: string[],
  dimensionName: string,
) => {
  const index = dashboard.visibleDimensions.indexOf(dimensionName);
  if (index !== -1) {
    dashboard.visibleDimensions.splice(index, 1);
  } else {
    dashboard.visibleDimensions.push(dimensionName);
  }

  dashboard.allDimensionsVisible =
    dashboard.visibleDimensions.length === allDimensions.length;
};

export const toggleAllDimensionsVisibility = (
  { dashboard }: DashboardMutables,
  allDimensions: string[],
) => {
  const allSelected =
    dashboard.visibleDimensions.length === allDimensions.length;

  dashboard.visibleDimensions = allSelected
    ? allDimensions.slice(0, 1)
    : [...allDimensions];
  dashboard.allDimensionsVisible = !dashboard.allDimensionsVisible;
};

export const setDimensionVisibility = (
  { dashboard }: DashboardMutables,
  dimensions: string[],
  allDimensions: string[],
) => {
  dashboard.visibleDimensions = [...dimensions];

  dashboard.allDimensionsVisible =
    dashboard.visibleDimensions.length === allDimensions.length;
};

export const dimensionActions = {
  /**
   * Sets the primary dimension for the dashboard, which
   * activates the dimension table. Setting the primary dimension
   * to undefined closes the dimension table.
   */
  setPrimaryDimension,
  toggleDimensionVisibility,
  toggleAllDimensionsVisibility,
  setDimensionVisibility,
};
