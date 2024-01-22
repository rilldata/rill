import { getDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
import { filterItemsSortFunction } from "@rilldata/web-common/features/dashboards/state-managers/selectors/filters";
import {
  createAndExpression,
  forEachExpression,
  getValuesInExpression,
  matchExpressionByName,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  MetricsViewSpecDimensionV2,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import type { AtLeast } from "../types";
import type { DashboardDataSources } from "./types";

export const getFiltersForOtherDimensions = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
): ((dimName: string) => V1Expression) => {
  return (dimName: string) => {
    const exprIdx = getWhereFilterExpressionIndex(dashData)(dimName);
    if (exprIdx === undefined || exprIdx === -1)
      return dashData.dashboard.whereFilter;

    return createAndExpression(
      dashData.dashboard.whereFilter.cond?.exprs?.filter(
        (e) => !matchExpressionByName(e, dimName),
      ) ?? [],
    );
  };
};

export const selectedDimensionValues = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
): ((dimName: string) => string[]) => {
  return (dimName: string) => {
    // FIXME: it is possible for this way of accessing the filters
    // to return the same value twice, which would seem to indicate
    // a bug in the way we're setting the filters / active values.
    // Need to investigate further to determine whether this is a
    // problem with the runtime or the client, but for now wrapping
    // it in a set dedupes the values.
    return [
      ...new Set(
        getValuesInExpression(
          getWhereFilterExpression(dashData)(dimName),
        ) as string[],
      ),
    ];
  };
};

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

export const dimensionHasFilter = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  return (dimName: string) => {
    return getWhereFilterExpression(dashData)(dimName) !== undefined;
  };
};

export const getWhereFilterExpression = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
): ((name: string) => V1Expression | undefined) => {
  return (name: string) =>
    dashData.dashboard.whereFilter.cond?.exprs?.find((e) =>
      matchExpressionByName(e, name),
    );
};

export const getWhereFilterExpressionIndex = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
): ((name: string) => number | undefined) => {
  return (name: string) =>
    dashData.dashboard.whereFilter?.cond?.exprs?.findIndex((e) =>
      matchExpressionByName(e, name),
    );
};

export type DimensionFilterItem = {
  name: string;
  label: string;
  selectedValues: string[];
};
export function getDimensionFilterItems(
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) {
  return (dimensionIdMap: Map<string, MetricsViewSpecDimensionV2>) => {
    if (!dashData.dashboard.whereFilter) return [];

    const filteredDimensions: DimensionFilterItem[] = [];
    const addedDimension = new Set<string>();
    forEachExpression(dashData.dashboard.whereFilter, (e) => {
      if (
        e.cond?.op !== V1Operation.OPERATION_IN &&
        e.cond?.op !== V1Operation.OPERATION_NIN
      ) {
        return;
      }
      const ident = e.cond?.exprs?.[0].ident;
      if (
        ident === undefined ||
        addedDimension.has(ident) ||
        !dimensionIdMap.has(ident)
      ) {
        return;
      }
      const dim = dimensionIdMap.get(ident);
      if (!dim) {
        return;
      }
      addedDimension.add(ident);
      filteredDimensions.push({
        name: ident,
        label: getDisplayName(dim),
        selectedValues: e.cond.exprs?.slice(1).map((e) => e.val) as any[],
      });
    });

    // sort based on name to make sure toggling include/exclude is not jarring
    return filteredDimensions.sort(filterItemsSortFunction);
  };
}

export const getAllDimensionFilterItems = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  return (
    dimensionFilterItem: DimensionFilterItem[],
    dimensionIdMap: Map<string, MetricsViewSpecDimensionV2>,
  ) => {
    const allDimensionFilterItem = [...dimensionFilterItem];

    // if the temporary filter is a dimension filter add it
    if (
      dashData.dashboard.temporaryFilterName &&
      dimensionIdMap.has(dashData.dashboard.temporaryFilterName)
    ) {
      allDimensionFilterItem.push({
        name: dashData.dashboard.temporaryFilterName,
        label: getDisplayName(
          dimensionIdMap.get(dashData.dashboard.temporaryFilterName),
        ),
        selectedValues: [],
      });
    }

    // sort based on name to make sure toggling include/exclude is not jarring
    return allDimensionFilterItem.sort(filterItemsSortFunction);
  };
};

export const unselectedDimensionValues = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  return (dimensionName: string, values: unknown[]): unknown[] => {
    const expr = getWhereFilterExpression(dashData)(dimensionName);
    if (expr === undefined) {
      return values;
    }

    return (
      expr.cond?.exprs
        ?.filter((e, i) => i > 0 && values.includes(e.val))
        .map((e) => e.val) ?? []
    );
  };
};

export const dimensionFilterSelectors = {
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

  /**
   * Check if a dimension has any filter
   */
  dimensionHasFilter,

  /**
   * Get filter items based on currently selected values for a dimension
   */
  getDimensionFilterItems,

  /**
   * Get filter items on dimension along with an empty entry for temporary filter if it is a dimension
   */
  getAllDimensionFilterItems,

  unselectedDimensionValues,
};
