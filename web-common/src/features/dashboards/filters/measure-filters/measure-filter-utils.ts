import {
  mapExprToMeasureFilter,
  mapMeasureFilterToExpr,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  createAndExpression,
  createSubQueryExpression,
  filterExpressions,
  isExpressionUnsupported,
  removeWrapperAndOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

export function mergeDimensionAndMeasureFilters(
  whereFilter: V1Expression | undefined,
  dimensionThresholdFilters: DimensionThresholdFilter[],
) {
  if (!whereFilter) return createAndExpression([]);
  const where =
    filterExpressions(whereFilter, () => true) ?? createAndExpression([]);
  where.cond?.exprs?.push(
    ...dimensionThresholdFilters.map(convertDimensionThresholdFilter),
  );
  return where;
}

/**
 * Splits where filter into dimension and measure filters.
 * Measure filters will be sub-queries
 */
export function splitWhereFilter(whereFilter: V1Expression | undefined) {
  if (whereFilter && isExpressionUnsupported(whereFilter)) {
    return { dimensionFilters: whereFilter, dimensionThresholdFilters: [] };
  }

  const dimensionFilters = createAndExpression([]);
  const dimensionThresholdFilters: DimensionThresholdFilter[] = [];
  whereFilter?.cond?.exprs?.filter((e) => {
    const subqueryExpr = e.cond?.exprs?.[1];

    // While all the types support multiple measure filters per dimension our UI doesn't allow this right now.
    // So unwrap while trying to validate a measure filter.
    const unwrappedHavingFilter = removeWrapperAndOrExpression(
      subqueryExpr?.subquery?.having,
    );
    const mappedMeasureFilter = mapExprToMeasureFilter(unwrappedHavingFilter);
    // If there is no valid measure filter at level one then we do not support it right now.
    if (!mappedMeasureFilter) {
      dimensionFilters.cond?.exprs?.push(e);
    } else {
      dimensionThresholdFilters.push({
        name: subqueryExpr?.subquery?.dimension ?? "",
        filters: [mappedMeasureFilter],
      });
    }
  });

  return { dimensionFilters, dimensionThresholdFilters };
}

function convertDimensionThresholdFilter(
  dtf: DimensionThresholdFilter,
): V1Expression {
  return createSubQueryExpression(
    dtf.name,
    dtf.filters.map((f) => f.measure),
    createAndExpression(
      dtf.filters.map(mapMeasureFilterToExpr).filter(Boolean) as V1Expression[],
    ),
  );
}
