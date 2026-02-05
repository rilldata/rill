import type { CartesianChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/CartesianChartProvider";
import type { ComboChartSpec } from "@rilldata/web-common/features/components/charts/combo/ComboChartProvider";
import type { HeatmapChartSpec } from "@rilldata/web-common/features/components/charts/heatmap/HeatmapChartProvider";
import type {
  ChartSortDirection,
  FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
import { isFieldConfig } from "@rilldata/web-common/features/components/charts/util";
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

/**
 * Converts a Vega-style sort configuration to Rill's aggregation sort format.
 */
export function vegaSortToAggregationSort(
  encoder: "x" | "y",
  config: CartesianChartSpec | HeatmapChartSpec | ComboChartSpec,
  defaultSort: ChartSortDirection,
): V1MetricsViewAggregationSort | undefined {
  const encoderConfig = config[encoder];

  if (!encoderConfig) {
    return undefined;
  }

  let sort = encoderConfig.sort;

  if (!sort || Array.isArray(sort)) {
    sort = defaultSort;
  }

  let field: string | undefined;
  let desc: boolean = false;

  switch (sort) {
    case "x":
    case "-x":
      field = config.x?.field;
      desc = sort === "-x";
      break;
    case "y":
    case "-y":
      if ("y" in config) {
        field = config.y?.field;
      } else if ("y1" in config) {
        field = config.y1?.field;
      }
      desc = sort === "-y";
      break;
    case "color":
    case "-color":
      field = isFieldConfig(config.color) ? config.color.field : undefined;
      desc = sort === "-color";
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
