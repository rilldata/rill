import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { VisualizationSpec } from "svelte-vega";
import type { ChartDataResult } from "../../../../components/charts/types";
import {
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../builder";
import type { HeatmapChartSpec } from "./HeatmapChart";

function createHeatmapSortEncoding(
  axisType: "x" | "y",
  config: HeatmapChartSpec,
  data: ChartDataResult,
) {
  const axisConfig = config[axisType];

  if (!axisConfig?.field || axisConfig.type !== "nominal") {
    return undefined;
  }
  // Use the pre-computed domain values from the query
  return data.domainValues?.[axisConfig.field];
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

  const xSort = createHeatmapSortEncoding("x", config, data);
  if (xSort !== undefined) {
    xEncoding.sort = xSort;
  }

  const ySort = createHeatmapSortEncoding("y", config, data);
  if (ySort !== undefined) {
    yEncoding.sort = ySort;
  }

  spec.encoding = {
    x: xEncoding,
    y: yEncoding,
    tooltip: createDefaultTooltipEncoding(
      [config.x, config.y, config.color],
      data,
    ),
  };

  spec.layer = [
    {
      mark: "rect",
      encoding: {
        color: createColorEncoding(config.color, data),
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
        color: {
          expr: `luminance ( scale ( 'color', datum['${sanitizeValueForVega(config.color?.field ?? "")}'] ) ) > 0.45 ? '#111827' : '#e5e7eb'`,
        },
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
          type:
            config.color?.type === "value"
              ? "nominal"
              : config.color?.type || "quantitative",
          ...(config.color?.type === "quantitative" &&
            config.color?.field && {
              formatType: sanitizeFieldName(config.color.field),
            }),
        },
      },
    });
  }

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
