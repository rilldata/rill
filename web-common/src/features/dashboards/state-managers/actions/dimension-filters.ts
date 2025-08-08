import { page } from "$app/stores";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  createInExpression,
  createLikeExpression,
  getValueIndexInExpression,
  getValuesInExpression,
  negateExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { getWhereFilterExpressionIndex } from "../selectors/dimension-filters";
import type { DashboardMutables } from "./types";
import { get } from "svelte/store";

export function toggleDimensionValueSelection(
  { dashboard }: DashboardMutables,
  dimensionName: string,
  dimensionValue: string,
  keepPillVisible?: boolean,
  /**
   * This marks the value as being exclusive. All other selected values will be unselected.
   */
  isExclusiveFilter?: boolean,
) {
  return toggleMultipleDimensionValueSelections(
    { dashboard },
    dimensionName,
    [dimensionValue],
    keepPillVisible,
    isExclusiveFilter,
  );
}

export function toggleMultipleDimensionValueSelections(
  { dashboard }: DashboardMutables,
  dimensionName: string,
  dimensionValues: string[],
  keepPillVisible?: boolean,
  isExclusiveFilter?: boolean,
) {
  if (dashboard.temporaryFilterName !== null) {
    dashboard.temporaryFilterName = null;
  }

  const isExclude = !!dashboard.dimensionFilterExcludeMode.get(dimensionName);
  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(dimensionName);
  if (exprIdx === undefined || exprIdx === -1) {
    dashboard.whereFilter.cond?.exprs?.push(
      createInExpression(dimensionName, dimensionValues, isExclude),
    );
    return;
  }

  const expr = dashboard.whereFilter.cond?.exprs?.[exprIdx];
  if (!expr?.cond?.exprs) {
    // should never happen since getWhereFilterExpressionIndex runs a find
    return;
  }

  const wasInListFilter =
    dashboard.dimensionsWithInlistFilter.includes(dimensionName);
  const wasLikeFilter =
    expr.cond?.op === V1Operation.OPERATION_LIKE ||
    expr.cond?.op === V1Operation.OPERATION_NLIKE;
  if (wasInListFilter || wasLikeFilter) {
    eventBus.emit("notification", {
      message: "Converted filter type to Select",
      link: {
        text: "Undo",
        href: get(page).url.href,
      },
    });
  }

  dashboard.dimensionsWithInlistFilter =
    dashboard.dimensionsWithInlistFilter.filter((d) => d !== dimensionName);
  if (wasLikeFilter) {
    eventBus.emit("notification", {
      message: "Converted filter type to Select",
      link: {
        text: "Undo",
        href: get(page).url.href,
      },
    });
    dashboard.whereFilter.cond!.exprs![exprIdx] = createInExpression(
      dimensionName,
      dimensionValues,
      isExclude,
    );
    return;
  }

  dimensionValues.forEach((v) => {
    const removedIndex = toggleDimensionFilterValue(
      expr,
      v,
      !!isExclusiveFilter,
    );
    if (removedIndex === -1) return;

    // Only decrement pinIndex if the removed value was before the pinned value
    if (dashboard.tdd.pinIndex >= removedIndex) {
      dashboard.tdd.pinIndex--;
    }
  });

  // remove the dimension entry if all values are removed
  if (expr.cond.exprs.length === 1) {
    dashboard.whereFilter.cond?.exprs?.splice(exprIdx, 1);
    if (keepPillVisible) {
      dashboard.temporaryFilterName = dimensionName;
    }
  }
}

export function applyDimensionInListMode(
  { dashboard }: DashboardMutables,
  dimensionName: string,
  values: string[],
) {
  if (dashboard.temporaryFilterName !== null) {
    dashboard.temporaryFilterName = null;
  }

  if (!dashboard.whereFilter.cond?.exprs) return;

  const isExclude = !!dashboard.dimensionFilterExcludeMode.get(dimensionName);
  const expr = createInExpression(dimensionName, values, isExclude);
  if (!dashboard.dimensionsWithInlistFilter.includes(dimensionName)) {
    dashboard.dimensionsWithInlistFilter.push(dimensionName);
  }
  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(dimensionName);
  if (exprIdx === undefined || exprIdx === -1) {
    dashboard.whereFilter.cond.exprs.push(expr);
  } else {
    dashboard.whereFilter.cond.exprs[exprIdx] = expr;
  }
}

export function applyDimensionContainsMode(
  { dashboard }: DashboardMutables,
  dimensionName: string,
  searchText: string,
) {
  if (dashboard.temporaryFilterName !== null) {
    dashboard.temporaryFilterName = null;
  }

  if (!dashboard.whereFilter.cond?.exprs) return;

  const isExclude = !!dashboard.dimensionFilterExcludeMode.get(dimensionName);
  const expr = createLikeExpression(
    dimensionName,
    `%${searchText}%`,
    isExclude,
  );
  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(dimensionName);
  if (exprIdx === undefined || exprIdx === -1) {
    dashboard.whereFilter.cond.exprs.push(expr);
  } else {
    dashboard.whereFilter.cond.exprs[exprIdx] = expr;
  }
}

export function toggleDimensionFilterMode(
  { dashboard }: DashboardMutables,
  dimensionName: string,
) {
  const exclude = dashboard.dimensionFilterExcludeMode.get(dimensionName);
  dashboard.dimensionFilterExcludeMode.set(dimensionName, !exclude);

  if (!dashboard.whereFilter?.cond?.exprs) {
    return;
  }

  const exprIdx = dashboard.whereFilter.cond.exprs.findIndex(
    (e) => e.cond?.exprs?.[0].ident === dimensionName,
  );
  if (exprIdx === -1) {
    return;
  }
  dashboard.whereFilter.cond.exprs[exprIdx] = negateExpression(
    dashboard.whereFilter.cond.exprs[exprIdx],
  );
}

export function removeDimensionFilter(
  { dashboard }: DashboardMutables,
  dimensionName: string,
) {
  if (dashboard.temporaryFilterName === dimensionName) {
    dashboard.temporaryFilterName = null;
    return;
  }

  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(dimensionName);
  if (exprIdx === undefined || exprIdx === -1) return;
  dashboard.whereFilter?.cond?.exprs?.splice(exprIdx, 1);
}

export function selectItemsInFilter(
  { dashboard }: DashboardMutables,
  dimensionName: string,
  values: (string | null)[],
) {
  const isExclude = !!dashboard.dimensionFilterExcludeMode.get(dimensionName);
  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(dimensionName);
  if (exprIdx === undefined || exprIdx === -1) {
    dashboard.whereFilter.cond?.exprs?.push(
      createInExpression(dimensionName, values, isExclude),
    );
    return;
  }

  const expr = dashboard.whereFilter.cond?.exprs?.[exprIdx];
  if (!expr?.cond?.exprs) {
    // should never happen since getWhereFilterExpressionIndex runs a find
    return;
  }

  // preserve old selections and add only new ones
  const oldValues = getValuesInExpression(expr);
  const newValues = values.filter((v) => !oldValues.includes(v));
  // newValuesSelected = newValues.length; // TODO
  expr.cond.exprs.push(...newValues.map((v): V1Expression => ({ val: v })));
}

export function deselectItemsInFilter(
  { dashboard }: DashboardMutables,
  dimensionName: string,
  values: (string | null)[],
) {
  const exprIdx = getWhereFilterExpressionIndex({ dashboard })(dimensionName);
  if (exprIdx === undefined || exprIdx === -1) {
    return;
  }

  const expr = dashboard.whereFilter.cond?.exprs?.[exprIdx];
  if (!expr?.cond?.exprs) {
    // should never happen since getWhereFilterExpressionIndex runs a find
    return;
  }

  // remove only deselected values
  const oldValues = getValuesInExpression(expr);
  const newValues = oldValues.filter((v) => !values.includes(v));

  if (newValues.length) {
    expr.cond.exprs.splice(
      1,
      expr.cond.exprs.length - 1,
      ...newValues.map((v): V1Expression => ({ val: v })),
    );
  } else {
    dashboard.whereFilter.cond?.exprs?.splice(exprIdx, 1);
  }
}

export function setFilters(
  { dashboard }: DashboardMutables,
  filter: V1Expression,
) {
  const { dimensionFilters, dimensionThresholdFilters } =
    splitWhereFilter(filter);
  dashboard.whereFilter = dimensionFilters;
  dashboard.dimensionThresholdFilters = dimensionThresholdFilters;
}

export function toggleDimensionFilterValue(
  expr: V1Expression,
  dimensionValue: string,
  isExclusiveFilter: boolean,
) {
  if (!expr.cond?.exprs) return -1;

  const inIdx = getValueIndexInExpression(expr, dimensionValue);
  if (inIdx === -1) {
    if (isExclusiveFilter) {
      expr.cond.exprs.splice(1, expr.cond.exprs.length - 1, {
        val: dimensionValue,
      });
    } else {
      expr.cond.exprs.push({ val: dimensionValue });
    }
    return -1;
  } else {
    expr.cond.exprs.splice(inIdx, 1);
    return inIdx;
  }
}

export const dimensionFilterActions = {
  /**
   * Toggles whether the given dimension value is selected in the
   * dimension filter for the given dimension.
   *
   * Note that this is different than the include/exclude mode for
   * dimension filters. This is a toggle for a specific value, whereas
   * the include/exclude mode is a toggle for the entire dimension.
   */
  toggleDimensionValueSelection,
  toggleMultipleDimensionValueSelections,
  applyDimensionInListMode,
  applyDimensionContainsMode,
  toggleDimensionFilterMode,
  removeDimensionFilter,
  selectItemsInFilter,
  deselectItemsInFilter,
  setFilters,
};
