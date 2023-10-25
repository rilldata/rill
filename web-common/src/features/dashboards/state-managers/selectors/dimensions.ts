import type { MetricsViewSpecDimensionV2 } from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";

export const allDimensions = ({
  metricsSpecQueryResult,
}: DashboardDataSources): MetricsViewSpecDimensionV2[] | undefined => {
  return metricsSpecQueryResult.data?.dimensions;
};

export const getDimensionByName = (
  args: DashboardDataSources
): ((name: string) => MetricsViewSpecDimensionV2 | undefined) => {
  return (name: string) => {
    return allDimensions(args)?.find((dimension) => dimension.name === name);
  };
};

export const getDimensionDisplayName = (
  args: DashboardDataSources
): ((name: string) => string) => {
  return (name: string) => {
    const dim = getDimensionByName(args)(name);
    return dim?.label || name;
  };
};

export const dimensionSelectors = {
  /**
   * Gets all dimensions for the dashboard, or undefined if there are none.
   */
  allDimensions,
  /**
   * Returns a function that can be used to get a MetricsViewSpecDimensionV2
   * by name; this fn returns undefined if the dashboard has no dimension with that name.
   */
  getDimensionByName,
  /**
   * Returns a function that can be used to get a dimension's display
   * given its "key" name.
   */
  getDimensionDisplayName,
};
