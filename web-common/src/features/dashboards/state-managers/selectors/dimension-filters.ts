import type {
  MetricsViewFilterCond,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./types";
import { getFiltersForOtherDimensions as getFiltersForOtherDimensionsUnconnected } from "../../selectors";
import type { AtLeast } from "../types";

export const getFiltersForOtherDimensions = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
): ((dimName: string) => V1MetricsViewFilter) => {
  return (dimName: string) =>
    getFiltersForOtherDimensionsUnconnected(
      dashData.dashboard.filters,
      dimName,
    );
};

export const selectedDimensionValues = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
): ((dimName: string) => string[]) => {
  return (dimName: string) => {
    const filterKey = filterModeKey(dashData)(dimName);

    // FIXME: it is possible for this way of accessing the filters
    // to return the same value twice, which would seem to indicate
    // a bug in the way we're setting the filters / active values.
    // Need to investigate further to determine whether this is a
    // problem with the runtime or the client, but for now wrapping
    // it in a set dedupes the values.
    return [
      ...new Set(
        (dashData.dashboard.filters[filterKey]?.find((d) => d.name === dimName)
          ?.in as string[]) ?? [],
      ),
    ];
  };
};

export const getAllFilters = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
): V1MetricsViewFilter => dashData.dashboard.filters;

export const atLeastOneSelection = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
): ((dimName: string) => boolean) => {
  return (dimName: string) =>
    selectedDimensionValues(dashData)(dimName).length > 0;
};

export const isFilterExcludeMode = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
): ((dimName: string) => boolean) => {
  return (dimName: string) =>
    dashData.dashboard.dimensionFilterExcludeMode.get(dimName) ?? false;
};

const filterModeKey = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
): ((dimName: string) => "exclude" | "include") => {
  return (dimName: string) =>
    isFilterExcludeMode(dashData)(dimName) ? "exclude" : "include";
};

export const filtersForCurrentExcludeMode = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
): ((dimName: string) => MetricsViewFilterCond[] | undefined) => {
  return (dimName: string) =>
    dashData.dashboard.filters[filterModeKey(dashData)(dimName)];
};

export const dimensionFilterSelectors = {
  /**
   * Gets all dimension filters for every dimension in the dashboard.
   */
  getAllFilters,
  /**
   * Returns a function that can be used to get
   * a copy of the dashboard's V1MetricsViewFilter that does not include
   * the filters for the specified dimension name.
   */
  getFiltersForOtherDimensions,

  /**
   * Returns a function that can be used to get the selected values
   * for the specified dimension name.
   */
  selectedDimensionValues,

  /**
   * Returns a function that can be used to get whether the specified
   * dimension has at least one selected value.
   */
  atLeastOneSelection,

  /**
   * Returns a function that can be used to get whether the specified
   * dimension is in exclude mode.
   */
  isFilterExcludeMode,
};
