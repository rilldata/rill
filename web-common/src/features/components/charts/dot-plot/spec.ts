import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
import {
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "@rilldata/web-common/features/components/charts/builder";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { LayerSpec } from "vega-lite/build/src/spec/layer";
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import type { DotPlotChartSpec } from "./DotPlotChartProvider";

export function generateVLDotPlotSpec(
  config: DotPlotChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  if (!config.y?.field || !config.x?.field) {
    throw new Error(
      "Dot plot requires both y (dimension) and x (measure) fields",
    );
  }

  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, config.color);

  const yField = sanitizeValueForVega(config.y.field);
  const xField = sanitizeValueForVega(config.x.field);
  const detailField = sanitizeValueForVega(config.detail?.field);
  const colorField =
    typeof config.color === "object" ? config.color.field : undefined;

  spec.encoding = {
    y: createPositionEncoding(config.y, data),
    x: createPositionEncoding(config.x, data),
  };

  const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [];

  const rangeLayer: UnitSpec<Field> = {
    mark: {
      type: "rect",
      cornerRadius: 4,
      opacity: 0.2,
      stroke: data.theme.primary.hex(),
      strokeWidth: 1,
      fill: data.theme.primary.hex(),
      height: { band: 0.75 },
    },
    transform: [
      {
        aggregate: [
          {
            op: "min",
            field: xField,
            as: "min_x",
          },
          {
            op: "max",
            field: xField,
            as: "max_x",
          },
          {
            op: "mean",
            field: xField,
            as: "avg_x",
          },
        ],
        groupby: [yField],
      },
    ],
    encoding: {
      y: {
        field: yField,
        type: "nominal",
        axis: { title: config.y?.showAxisTitle ? yField : null },
      },
      x: {
        field: "min_x",
        type: "quantitative",
        axis: { title: config.x?.showAxisTitle ? xField : null },
        scale: { zero: false },
      },
      x2: {
        field: "max_x",
        type: "quantitative",
      },
      tooltip: [
        {
          title: data.fields[yField]?.displayName || yField,
          field: yField,
          type: "nominal",
        },
        {
          title: "Min",
          field: "min_x",
          type: "quantitative",
          formatType: sanitizeFieldName(xField),
        },
        {
          title: "Max",
          field: "max_x",
          type: "quantitative",
          formatType: sanitizeFieldName(xField),
        },
        {
          title: "Avg",
          field: "avg_x",
          type: "quantitative",
          formatType: sanitizeFieldName(xField),
        },
      ],
    },
  };

  const jitterEnabled = config.jitter === true;

  const dotsLayerTransforms = [
    ...(jitterEnabled
      ? [
          {
            calculate: "(random() - 0.5) * 0.4",
            as: "jitter_y",
          },
        ]
      : []),
    ...(!detailField
      ? [
          {
            calculate: `datum.${yField} + '_' + datum.${xField} + '_' + (datum.row_number || random())`,
            as: "detail_id",
          },
        ]
      : []),
  ];

  const dotsLayer: UnitSpec<Field> = {
    mark: {
      type: "circle",
      filled: true,
      size: 60,
      opacity: 0.7,
    },
    ...(dotsLayerTransforms.length > 0 && { transform: dotsLayerTransforms }),
    encoding: {
      y: {
        field: yField,
        type: "nominal",
        axis: { title: config.y?.showAxisTitle ? yField : null },
      },
      ...(jitterEnabled && {
        yOffset: {
          field: "jitter_y",
          type: "quantitative",
        },
      }),
      x: {
        field: xField,
        type: "quantitative",
        axis: { title: config.x?.showAxisTitle ? xField : null },
        scale: { zero: false },
      },
      detail: detailField
        ? {
            field: detailField,
            type: "nominal",
          }
        : {
            field: "detail_id",
            type: "nominal",
          },
      ...(colorField && {
        color: createColorEncoding(config.color, data),
      }),
      tooltip: (() => {
        const tooltipFields = createDefaultTooltipEncoding(
          [config.y, config.x, config.detail, config.color].filter(Boolean),
          data,
        );
        if (tooltipFields.length > 0) {
          tooltipFields[0] = { ...tooltipFields[0], title: undefined };
        }
        return tooltipFields;
      })(),
    },
  };

  const avgLayer: UnitSpec<Field> = {
    mark: {
      type: "tick",
      stroke: data.theme.primary.hex(),
      strokeWidth: 2,
      opacity: 0.8,
      height: { band: 0.75 },
    },
    transform: [
      {
        aggregate: [
          {
            op: "mean",
            field: xField,
            as: "mean_x",
          },
        ],
        groupby: [yField],
      },
    ],
    encoding: {
      y: {
        field: yField,
        type: "nominal",
      },
      x: {
        field: "mean_x",
        type: "quantitative",
      },
    },
  };

  layers.push(rangeLayer, avgLayer, dotsLayer);

  spec.layer = layers;
  spec.config = vegaConfig;

  return spec;
}
