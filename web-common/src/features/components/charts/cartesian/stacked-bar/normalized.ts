import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
import {
  buildHoverRuleLayer,
  createCartesianMultiValueTooltipChannel,
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "@rilldata/web-common/features/components/charts/builder";
import {
  createComparisonOpacityEncoding,
  createComparisonTransforms,
  createComparisonXOffsetEncoding,
  MeasureKeyField,
} from "@rilldata/web-common/features/components/charts/comparison-builder";
import type { TooltipValue } from "@rilldata/web-common/features/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { LayerSpec } from "vega-lite/build/src/spec/layer";
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import type { Transform } from "vega-lite/build/src/transform";
import type { CartesianChartSpec } from "../CartesianChartProvider";
import { createVegaTransformPivotConfig } from "../util";

export function generateVLStackedBarNormalizedSpec(
  config: CartesianChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createMultiLayerBaseSpec();
  const baseEncoding = createEncoding(config, data);
  const vegaConfig = createConfigWithLegend(config, config.color);

  const colorField =
    typeof config.color === "object" ? config.color.field : undefined;
  const xField = sanitizeValueForVega(config.x?.field);
  const yField = sanitizeValueForVega(config.y?.field);

  // Check if comparison mode is enabled
  const hasComparison = data.hasComparison;

  if (baseEncoding.y && config.y?.field) {
    const yFieldRaw = config.y.field;

    baseEncoding.y = {
      ...baseEncoding.y,
      stack: "normalize",
      scale: {
        zero: false,
        // Add padding at the top for hover space since normalized charts go to 100%
        domainMax: 1.05,
      },
      axis: {
        ...(!config.y.showAxisTitle && { title: null }),
        format: ".0%",
      },
    };

    // Add a transform to calculate the percentage
    // When in comparison mode, include the measure_key in groupby
    // so each period (current vs comparison) totals to 100% independently
    const groupbyFields = config.x?.field ? [config.x.field] : [];
    if (hasComparison) {
      groupbyFields.push(MeasureKeyField);
    }

    const percentageTransforms: Transform[] = [
      {
        joinaggregate: [
          {
            op: "sum",
            field: yFieldRaw,
            as: "total",
          },
        ],
        groupby: groupbyFields,
      },
      {
        calculate: `datum['${yFieldRaw}'] / datum.total`,
        as: "percentage",
      },
    ];

    // Apply comparison transforms if needed
    if (hasComparison && colorField) {
      const comparisonTransforms = createComparisonTransforms(
        config.x?.field,
        config.y?.field,
        colorField,
      );
      spec.transform = [...comparisonTransforms, ...percentageTransforms];
    } else if (hasComparison) {
      const comparisonTransforms = createComparisonTransforms(
        config.x?.field,
        config.y?.field,
      );
      spec.transform = [...comparisonTransforms, ...percentageTransforms];
    } else {
      spec.transform = percentageTransforms;
    }

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

  const multiValueTooltipChannel = createCartesianMultiValueTooltipChannel(
    { x: config.x, colorField, yField },
    data,
  );

  spec.encoding = { x: createPositionEncoding(config.x, data) };

  const hoverRuleLayer = buildHoverRuleLayer({
    xField,
    domainValues: data.domainValues,
    isBarMark: true,
    defaultTooltip: baseEncoding.tooltip as TooltipValue[],
    multiValueTooltipChannel,
    xSort: config.x?.sort,
    primaryColor: data.theme.primary,
    isDarkMode: data.isDarkMode,
    xBand: config.x?.type === "temporal" ? 0.5 : undefined,
    pivot: createVegaTransformPivotConfig(
      xField,
      yField,
      colorField,
      !!hasComparison,
      !!multiValueTooltipChannel?.length,
    ),
  });

  const barLayer: UnitSpec<Field> = {
    mark: { type: "bar", clip: true, width: { band: 0.9 } },
    encoding: {
      y: baseEncoding.y,
      color: createColorEncoding(config.color, data),
      tooltip: baseEncoding.tooltip,
    },
  };

  if (hasComparison && colorField) {
    // Comparison mode for stacked bars: use transforms with color dimension
    barLayer.encoding!.xOffset = createComparisonXOffsetEncoding();
    barLayer.encoding!.opacity = createComparisonOpacityEncoding(yField);
  } else if (hasComparison) {
    barLayer.encoding!.xOffset = createComparisonXOffsetEncoding();
    barLayer.encoding!.opacity = createComparisonOpacityEncoding(yField);
  }

  const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [
    hoverRuleLayer,
    barLayer,
  ];

  spec.layer = layers;

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
