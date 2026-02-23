import type { CartesianChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/CartesianChartProvider";
import type { ComboChartSpec } from "@rilldata/web-common/features/components/charts/combo/ComboChartProvider";
import type { HeatmapChartSpec } from "@rilldata/web-common/features/components/charts/heatmap/HeatmapChartProvider";
import {
  ChartSortType,
  type ChartSortDirection,
  type FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
import { isFieldConfig } from "@rilldata/web-common/features/components/charts/util";
import { ComparisonDeltaAbsoluteSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  V1Expression,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";

export function getFilterWithNullHandling(
  where: V1Expression | undefined,
  fieldConfig: FieldConfig | undefined,
): V1Expression | undefined {
  if (!fieldConfig || fieldConfig.showNull || fieldConfig.type !== "nominal") {
    return where;
  }

  const excludeNullFilter = createInExpression(fieldConfig.field, [null], true);
  return mergeFilters(where, excludeNullFilter);
}

export function isSortByDelta(sort: ChartSortDirection | undefined) {
  if (!sort) return false;
  return (
    sort === ChartSortType.Y_DELTA_ASC || sort === ChartSortType.Y_DELTA_DESC
  );
}

/**
 * Converts a Vega-style sort configuration to Rill's aggregation sort format.
 */
export function vegaSortToAggregationSort(
  encoder: "x" | "y",
  config: CartesianChartSpec | HeatmapChartSpec | ComboChartSpec,
  defaultSort: ChartSortDirection,
  isComparisonActive = false,
): V1MetricsViewAggregationSort | undefined {
  const encoderConfig = config[encoder];

  if (!encoderConfig) {
    return undefined;
  }

  let sort = encoderConfig.sort;

  if (!sort || Array.isArray(sort)) {
    sort = defaultSort;
  }

  if (!isComparisonActive) {
    if (sort === ChartSortType.Y_DELTA_ASC) sort = ChartSortType.Y_ASC;
    else if (sort === ChartSortType.Y_DELTA_DESC) sort = ChartSortType.Y_DESC;
  }

  let field: string | undefined;
  let desc: boolean = false;

  switch (sort) {
    case ChartSortType.X_ASC:
    case ChartSortType.X_DESC:
      field = config.x?.field;
      desc = sort === ChartSortType.X_DESC;
      break;
    case ChartSortType.Y_ASC:
    case ChartSortType.Y_DESC:
      if ("y" in config) {
        field = config.y?.field;
      } else if ("y1" in config) {
        field = config.y1?.field;
      }
      desc = sort === ChartSortType.Y_DESC;
      break;
    case ChartSortType.Y_DELTA_ASC:
    case ChartSortType.Y_DELTA_DESC:
      if ("y" in config) {
        field = config.y?.field;
      } else if ("y1" in config) {
        field = config.y1?.field;
      }
      if (field) {
        field = field + ComparisonDeltaAbsoluteSuffix;
      }
      desc = sort === ChartSortType.Y_DELTA_DESC;
      break;
    case ChartSortType.COLOR_ASC:
    case ChartSortType.COLOR_DESC:
      field = isFieldConfig(config.color) ? config.color.field : undefined;
      desc = sort === ChartSortType.COLOR_DESC;
      break;
    default:
      return undefined;
  }

  if (!field) return undefined;

  return {
    name: field,
    desc,
  };
}
