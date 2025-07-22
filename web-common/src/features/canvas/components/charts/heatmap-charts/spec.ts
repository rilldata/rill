import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import type { VisualizationSpec } from "svelte-vega";
import {
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../builder";
import type { ChartDataResult } from "../types";
import type { HeatmapChartSpec } from "./HeatmapChart";

function createHeatmapSortEncoding(
  axisType: "x" | "y",
  config: HeatmapChartSpec,
) {
  const axisConfig = config[axisType];

  if (!axisConfig || axisConfig.type !== "nominal") {
    return undefined;
  }

  const sort = axisConfig.sort;

  if (Array.isArray(sort)) {
    return sort;
  }
  if (!sort) {
    return undefined;
  }

  let sortField: string | undefined;
  let sortOrder: "ascending" | "descending" = "descending";

  switch (sort) {
    case "x":
    case "-x":
      sortField = config.x?.field;
      sortOrder = sort === "-x" ? "descending" : "ascending";
      break;
    case "y":
    case "-y":
      sortField = config.y?.field;
      sortOrder = sort === "-y" ? "descending" : "ascending";
      break;
    case "color":
    case "-color":
      sortField = config.color?.field;
      sortOrder = sort === "-color" ? "descending" : "ascending";
      break;
    default:
      return undefined;
  }

  if (!sortField) {
    return undefined;
  }

  if (sort === "color" || sort === "-color") {
    return {
      op: "sum" as const,
      field: sortField,
      order: sortOrder,
    };
  }

  return {
    field: sortField,
    order: sortOrder,
  };
}

export function generateVLHeatmapSpec(
  config: HeatmapChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createMultiLayerBaseSpec();

  spec.description = "A heatmap chart with embedded data";

  const vegaConfig = createConfigWithLegend(
    config,
    config.color,
    {
      axis: { grid: true, tickBand: "extent" },
      axisX: {
        grid: true,
        gridDash: [],
        tickBand: "extent",
      },
      axisTemporal: { grid: true, zindex: 1 },
    },
    "right",
  );

  spec.height = "container";

  const xEncoding = createPositionEncoding(config.x, data);
  const yEncoding = createPositionEncoding(config.y, data);

  const xSort = createHeatmapSortEncoding("x", config);
  if (xSort !== undefined) {
    xEncoding.sort = xSort;
  }

  const ySort = createHeatmapSortEncoding("y", config);
  if (ySort !== undefined) {
    yEncoding.sort = ySort;
  }

  // Add transform to calculate threshold for text color - using 75th percentile for better contrast
  if (config.color?.field) {
    spec.transform = [
      {
        joinaggregate: [
          {
            op: "q3",
            field: config.color.field,
            as: "q3_value",
          },
          {
            op: "max",
            field: config.color.field,
            as: "max_value",
          },
          {
            op: "min",
            field: config.color.field,
            as: "min_value",
          },
        ],
      },
      {
        // Use white text only when value is in the top 25% and significantly above the 75th percentile
        calculate: `datum['${config.color.field}'] > datum.q3_value && (datum['${config.color.field}'] - datum.min_value) / (datum.max_value - datum.min_value) > 0.7`,
        as: "use_white_text",
      },
    ];
  }

  spec.encoding = {
    x: xEncoding,
    y: yEncoding,
  };

  spec.layer = [
    {
      mark: "rect",
      encoding: {
        color: createColorEncoding(config.color, data),
        tooltip: createDefaultTooltipEncoding(
          [config.x, config.y, config.color],
          data,
        ),
      },
    },
  ];

  if (config.show_data_labels === true) {
    spec.layer.push({
      mark: {
        type: "text",
        fontSize: 11,
        fontWeight: "normal",
        opacity: 0.9,
      },
      encoding: {
        // Use centered positioning for text on continuous scales
        x: {
          ...xEncoding,
          ...(config.x?.type === "temporal" && {
            bandPosition: 0.5,
          }),
        },

        text: {
          field: config.color?.field ? config.color.field : undefined,
          type: config.color?.type || "quantitative",
          ...(config.color?.type === "quantitative" &&
            config.color?.field && {
              formatType: sanitizeFieldName(config.color.field),
            }),
        },
        color: {
          value: "#111827",
          condition: {
            test: "datum.use_white_text",
            value: "#e5e7eb",
          },
        },
      },
    });
  }

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
