import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
import {
  buildHoverRuleLayer,
  createCartesianMultiValueTooltipChannel,
  createComparisonOpacityEncoding,
  createComparisonTransforms,
  createComparisonXOffsetEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "@rilldata/web-common/features/components/charts/builder";
import type { TooltipValue } from "@rilldata/web-common/features/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { LayerSpec } from "vega-lite/build/src/spec/layer";
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import type { CartesianChartSpec } from "../CartesianChartProvider";

export function generateVLStackedBarNormalizedSpec(
  config: CartesianChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createMultiLayerBaseSpec();

  // Check if comparison mode is enabled
  const hasComparison = data.hasComparison;
  const comparisonColorField =
    typeof config.color === "object" ? config.color.field : undefined;

  // For comparison mode with normalized stacks, we disable comparison
  // as normalized stacks with comparison would be confusing
  if (hasComparison && comparisonColorField) {
    // Fall back to regular stacked bar with comparison
    const transforms = createComparisonTransforms(
      config.x?.field,
      config.y?.field,
      comparisonColorField,
    );

    spec.transform = transforms;
    spec.encoding = {
      x: createPositionEncoding(config.x, data),
    };

    const barLayer: UnitSpec<Field> = {
      mark: { type: "bar", clip: true, width: { band: 1 } },
      encoding: {
        y: {
          field: "measure_value",
          type: "quantitative",
          title:
            data.fields[config.y?.field || ""]?.displayName || config.y?.field,
        },
        color: {
          field: sanitizeValueForVega(comparisonColorField),
          type: "nominal",
          title:
            data.fields[comparisonColorField]?.displayName ||
            comparisonColorField,
          legend: null,
        },
        opacity: createComparisonOpacityEncoding(config.y?.field || ""),
        xOffset: createComparisonXOffsetEncoding(comparisonColorField),
        tooltip: [
          {
            field: config.x?.field,
            type: config.x?.type === "temporal" ? "temporal" : "nominal",
            title:
              data.fields[config.x?.field || ""]?.displayName ||
              config.x?.field,
            ...(config.x?.type === "temporal" && { format: "%b %d, %Y %H:%M" }),
          },
          {
            field: "measure_value",
            type: "quantitative",
            title:
              data.fields[config.y?.field || ""]?.displayName ||
              config.y?.field,
            formatType: sanitizeFieldName(config.y?.field || ""),
          },
          {
            field: sanitizeValueForVega(comparisonColorField),
            type: "nominal",
            title:
              data.fields[comparisonColorField]?.displayName ||
              comparisonColorField,
          },
        ],
      },
      params: [
        {
          name: "hover",
          select: {
            type: "point",
            on: "pointerover",
            clear: "pointerout",
            encodings: ["x", "xOffset", "color"],
          },
        },
      ],
    };

    spec.layer = [barLayer];

    return {
      ...spec,
      ...(createConfigWithLegend(config, config.color) && {
        config: createConfigWithLegend(config, config.color),
      }),
    };
  }

  // Normal normalized mode without comparison
  const baseEncoding = createEncoding(config, data);
  const vegaConfig = createConfigWithLegend(config, config.color);

  if (baseEncoding.y && config.y?.field) {
    const yField = config.y.field;

    baseEncoding.y = {
      ...baseEncoding.y,
      stack: "normalize",
      ...(baseEncoding.y && {
        scale: {
          zero: false,
        },
      }),
      axis: {
        ...(!config.y.showAxisTitle && { title: null }),
        format: ".0%",
      },
    };

    // Add a transform to calculate the percentage
    spec.transform = [
      {
        joinaggregate: [
          {
            op: "sum",
            field: yField,
            as: "total",
          },
        ],
        groupby: config.x?.field ? [config.x.field] : [],
      },
      {
        calculate: `datum['${yField}'] / datum.total`,
        as: "percentage",
      },
    ];

    // Add percentage to tooltip
    const tooltipValues = createDefaultTooltipEncoding(
      [config.x, config.y, config.color],
      data,
    );
    baseEncoding.tooltip = tooltipValues
      .map((t: TooltipValue) => {
        if (t.field === yField) {
          return [
            {
              ...t,
            },
            {
              ...t,
              title: `${t.title} (%)`,
              field: "percentage",
              formatType: undefined,
              format: ".1%",
            },
          ];
        }
        return t;
      })
      .flat();
  }

  const colorField =
    typeof config.color === "object" ? config.color.field : undefined;
  const xField = sanitizeValueForVega(config.x?.field);
  const yField = sanitizeValueForVega(config.y?.field);

  const multiValueTooltipChannel = createCartesianMultiValueTooltipChannel(
    { x: config.x, colorField, yField },
    data,
  );

  spec.encoding = { x: createPositionEncoding(config.x, data) };

  const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [
    buildHoverRuleLayer({
      xField,
      domainValues: data.domainValues,
      isBarMark: true,
      defaultTooltip: baseEncoding.tooltip as TooltipValue[],
      multiValueTooltipChannel,
      xSort: config.x?.sort,
      primaryColor: data.theme.primary,
      isDarkMode: data.isDarkMode,
      xBand: config.x?.type === "temporal" ? 0.5 : undefined,
      pivot:
        xField && yField && colorField && multiValueTooltipChannel?.length
          ? { field: colorField, value: yField, groupby: [xField] }
          : undefined,
    }),
    {
      mark: { type: "bar", clip: true, width: { band: 0.9 } },
      encoding: {
        ...baseEncoding,
      },
    },
  ];

  spec.layer = layers;

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
