import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-filter-mode";
import { useDimensionSearch } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimensionFilterValues";
import { getDimensionDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
import { filterItemsSortFunction } from "@rilldata/web-common/features/dashboards/state-managers/selectors/filters";
import {
  forEachIdentifier,
  getValuesInExpression,
  isExpressionUnsupported,
  matchExpressionByName,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  MetricsViewSpecDimensionV2,
  V1Expression,
} from "@rilldata/web-common/runtime-client";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import { readable } from "svelte/store";
import type { AtLeast } from "../types";
import type { DashboardDataSources } from "./types";

export const selectedDimensionValues = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
): ((dimName: string) => string[]) => {
  return (dimName: string) => {
    // if it is a complex filter unsupported by UI then no values are selected
    if (isExpressionUnsupported(dashData.dashboard.whereFilter)) return [];

    const dimExpr = getWhereFilterExpression(dashData)(dimName);
    if (
      dimExpr?.cond?.op &&
      (dimExpr.cond.op === V1Operation.OPERATION_LIKE ||
        dimExpr.cond.op === V1Operation.OPERATION_NLIKE)
    )
      return [];

    // FIXME: it is possible for this way of accessing the filters
    // to return the same value twice, which would seem to indicate
    // a bug in the way we're setting the filters / active values.
    // Need to investigate further to determine whether this is a
    // problem with the runtime or the client, but for now wrapping
    // it in a set dedupes the values.
    return [...new Set(getValuesInExpression(dimExpr) as string[])];
  };
};

export const selectedDimensionValuesV2 = (
  instanceId: string,
  metricsViewNames: string[],
  whereFilter: V1Expression | undefined,
  dimensionName: string,
  timeStart?: string,
  timeEnd?: string,
): CompoundQueryResult<string[]> => {
  // if it is a complex filter unsupported by UI then no values are selected
  if (!whereFilter || isExpressionUnsupported(whereFilter))
    return readable({
      isFetching: false,
      isLoading: false,
      error: undefined,
      data: [],
    });

  const dimExpr = whereFilter.cond?.exprs?.find((e) =>
    matchExpressionByName(e, dimensionName),
  );
  if (!dimExpr?.cond?.op)
    return readable({
      isFetching: false,
      isLoading: false,
      error: undefined,
      data: [],
    });

  if (
    dimExpr.cond.op === V1Operation.OPERATION_IN ||
    dimExpr.cond.op === V1Operation.OPERATION_NIN
  ) {
    return readable({
      isFetching: false,
      isLoading: false,
      error: undefined,
      data: [...new Set(getValuesInExpression(dimExpr) as string[])],
    });
  }

  if (
    dimExpr.cond.op === V1Operation.OPERATION_LIKE ||
    dimExpr.cond.op === V1Operation.OPERATION_NLIKE
  ) {
    return useDimensionSearch(instanceId, metricsViewNames, dimensionName, {
      mode: DimensionFilterMode.Contains,
      searchText: (dimExpr.cond?.exprs?.[1]?.val as string) ?? "",
      values: [],
      timeStart,
      timeEnd,
      enabled: true,
    });
  }

  return readable({
    isFetching: false,
    isLoading: false,
    error: undefined,
    data: [],
  });
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
  mode: DimensionFilterMode;
  selectedValues: string[];
  inputText?: string;
  isInclude: boolean;
  metricsViewNames?: string[];
};
export function getDimensionFilterItems(
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) {
  return (dimensionIdMap: Map<string, MetricsViewSpecDimensionV2>) => {
    return getDimensionFilters(
      dimensionIdMap,
      dashData.dashboard.whereFilter,
      dashData.dashboard.dimensionsWithInlistFilter,
    );
  };
}

export function getDimensionFilters(
  dimensionIdMap: Map<string, MetricsViewSpecDimensionV2>,
  filter: V1Expression | undefined,
  dimensionsWithInlistFilter: string[],
) {
  if (!filter) return [];
  const filteredDimensions: DimensionFilterItem[] = [];
  const addedDimension = new Set<string>();
  forEachIdentifier(filter, (e, ident) => {
    if (addedDimension.has(ident) || !dimensionIdMap.has(ident)) return;
    const dim = dimensionIdMap.get(ident);
    if (!dim) {
      return;
    }
    addedDimension.add(ident);

    const op = e.cond?.op;
    if (op === V1Operation.OPERATION_IN || op === V1Operation.OPERATION_NIN) {
      const isInListMode = dimensionsWithInlistFilter.includes(ident);
      filteredDimensions.push({
        name: ident,
        label: getDimensionDisplayName(dim),
        mode: isInListMode
          ? DimensionFilterMode.InList
          : DimensionFilterMode.Select,
        selectedValues: getValuesInExpression(e),
        isInclude: e.cond?.op === V1Operation.OPERATION_IN,
      });
    } else if (
      op === V1Operation.OPERATION_LIKE ||
      op === V1Operation.OPERATION_NLIKE
    ) {
      filteredDimensions.push({
        name: ident,
        label: getDimensionDisplayName(dim),
        mode: DimensionFilterMode.Contains,
        selectedValues: [],
        inputText: e.cond?.exprs?.[1]?.val?.toString?.() ?? "",
        isInclude: e.cond?.op === V1Operation.OPERATION_IN,
      });
    }
  });

  // sort based on name to make sure toggling include/exclude is not jarring
  return filteredDimensions.sort(filterItemsSortFunction);
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
        label: getDimensionDisplayName(
          dimensionIdMap.get(dashData.dashboard.temporaryFilterName),
        ),
        mode: DimensionFilterMode.Select,
        selectedValues: [],
        isInclude: true,
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

    return values.filter(
      (v) => expr.cond?.exprs?.findIndex((e) => e.val === v) === -1,
    );
  };
};

export const includedDimensionValues = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  return (dimensionName: string): unknown[] => {
    const expr = getWhereFilterExpression(dashData)(dimensionName);
    if (expr === undefined || expr.cond?.op !== V1Operation.OPERATION_IN) {
      return [];
    }

    return getValuesInExpression(expr);
  };
};

export const hasAtLeastOneDimensionFilter = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  const whereFilter = dashData.dashboard.whereFilter;
  return whereFilter.cond?.exprs?.length && whereFilter.cond.exprs.length > 0;
};

export const dimensionFilterSelectors = {
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
  includedDimensionValues,
  hasAtLeastOneDimensionFilter,
};
