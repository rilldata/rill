import type { MetricsViewSpecDimension } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";

export const allDimensions = ({
  validMetricsView,
  validExplore,
}: Pick<
  DashboardDataSources,
  "validMetricsView" | "validExplore"
>): MetricsViewSpecDimension[] => {
  if (!validMetricsView?.dimensions || !validExplore?.dimensions) return [];

  return (
    validMetricsView.dimensions
      .filter(
        (d) =>
          validExplore.dimensions!.includes(d.name!) &&
          d.type !== "DIMENSION_TYPE_TIME",
      ) //
      // Sort the filtered dimensions based on their order in validExplore.dimensions
      .sort(
        (a, b) =>
          validExplore.dimensions!.indexOf(a.name!) -
          validExplore.dimensions!.indexOf(b.name!),
      )
  );
};

export const visibleDimensions = ({
  validMetricsView,
  validExplore,
  dashboard,
}: DashboardDataSources): MetricsViewSpecDimension[] => {
  if (
    !validMetricsView?.dimensions ||
    !validExplore?.dimensions ||
    !dashboard?.visibleDimensions ||
    !Array.isArray(dashboard.visibleDimensions)
  ) {
    return [];
  }

  const dimensions = validMetricsView.dimensions || [];

  return dashboard.visibleDimensions
    .map((dim) => dimensions.find((d) => d.name === dim))
    .filter(
      (d) => d && d?.type !== "DIMENSION_TYPE_TIME",
    ) as MetricsViewSpecDimension[];
};

export const dimensionTableColumnName = (
  dashData: DashboardDataSources,
): ((name: string) => string) => {
  return (name: string) => {
    const dim = getDimensionByName(dashData)(name);
    return dim?.column || name;
  };
};

export const getDimensionByName = (
  dashData: DashboardDataSources,
): ((name: string) => MetricsViewSpecDimension | undefined) => {
  return (name: string) => {
    return allDimensions(dashData)?.find(
      (dimension) => dimension.name === name,
    );
  };
};

export const getDimensionDisplayName = (
  dashData: DashboardDataSources,
): ((name: string) => string) => {
  return (name: string) => {
    const dim = getDimensionByName(dashData)(name);
    return (dim?.displayName?.length ? dim?.displayName : dim?.name) ?? name;
  };
};

export const comparisonDimension = (dashData: DashboardDataSources) => {
  const dimName = dashData.dashboard.selectedComparisonDimension;
  if (!dimName) return undefined;
  return getDimensionByName(dashData)(dimName);
};

export const dimensionSelectors = {
  /**
   * Gets all dimensions for the dashboard, or undefined if there are none.
   */
  allDimensions,

  /**
   * Gets all visible dimensions in the dashboard.
   */
  visibleDimensions,

  /**
   * Returns a function that can be used to get a MetricsViewSpecDimension
   * by name; this fn returns undefined if the dashboard has no dimension with that name.
   */
  getDimensionByName,
  /**
   * Returns a function that can be used to get a dimension's display
   * given its "key" name.
   */
  getDimensionDisplayName,

  /**
   * Gets the dimension that is currently being compared.
   * Returns undefined otherwise.
   */
  comparisonDimension,

  /**
   * Gets the name of the column that is currently selected in the dimension table.
   */
  dimensionTableColumnName,
};
