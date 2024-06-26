import { mapMeasureFilterToExpr } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  createAndExpression,
  filterExpressions,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  DimensionThresholdFilter,
  MetricsExplorerEntity,
} from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { V1Expression, V1Operation } from "@rilldata/web-common/runtime-client";

export const mergeMeasureFilters = (
  dashboard: MetricsExplorerEntity,
  whereFilter = dashboard.whereFilter,
) => {
  const where =
    filterExpressions(whereFilter, () => true) ?? createAndExpression([]);
  where.cond?.exprs?.push(
    ...dashboard.dimensionThresholdFilters.map(convertDimensionThresholdFilter),
  );
  return where;
};

const convertDimensionThresholdFilter = (
  dtf: DimensionThresholdFilter,
): V1Expression => {
  return {
    cond: {
      op: V1Operation.OPERATION_IN,
      exprs: [
        { ident: dtf.name },
        {
          subquery: {
            dimension: dtf.name,
            measures: dtf.filters.map((f) => f.measure),
            where: undefined,
            having: createAndExpression(
              dtf.filters
                .map(mapMeasureFilterToExpr)
                .filter(Boolean) as V1Expression[],
            ),
          },
        },
      ],
    },
  };
};
