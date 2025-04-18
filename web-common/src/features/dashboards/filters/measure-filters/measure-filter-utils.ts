import {
  mapExprToMeasureFilter,
  mapMeasureFilterToExpr,
  type MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  createAndExpression,
  createSubQueryExpression,
  filterExpressions,
  isExpressionUnsupported,
  isJoinerExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
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
    if (subqueryExpr?.subquery) {
      const filters = isJoinerExpression(subqueryExpr.subquery.having)
        ? (subqueryExpr.subquery.having?.cond?.exprs
            ?.map(mapExprToMeasureFilter)
            .filter(Boolean) as MeasureFilterEntry[])
        : [mapExprToMeasureFilter(subqueryExpr.subquery.having)];

      dimensionThresholdFilters.push({
        name: subqueryExpr.subquery.dimension ?? "",
        filters: filters.filter(Boolean) as MeasureFilterEntry[],
      });
      return;
    }

    dimensionFilters.cond?.exprs?.push(e);
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
