import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
import {
  buildHoverRuleLayer,
  createCartesianMultiValueTooltipChannel,
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
  createStackOverride,
} from "@rilldata/web-common/features/components/charts/builder";
import {
  createComparisonOpacityEncoding,
  createComparisonTransforms,
  createComparisonXOffsetEncoding,
} from "@rilldata/web-common/features/components/charts/comparison-builder";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/types_unstable/channeldef.js";
import type { UnitSpec } from "vega-lite/types_unstable/spec/unit.js";
import { type CartesianChartSpec } from "../CartesianChartProvider";
import {
  isHorizontal,
  toVerticalSpec,
  transposeCartesianSpec,
} from "../orientation";
import { createVegaTransformPivotConfig } from "../util";

export function generateVLStackedBarChartSpec(
  chartConfig: CartesianChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  // The spec is built for the vertical layout and transposed at the end when
  // the measure sits on x.
  const horizontal = isHorizontal(chartConfig);
  const config = toVerticalSpec(chartConfig);

  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, config.color);

  const colorField =
    typeof config.color === "object" ? config.color.field : undefined;
  const sanitizedXField = sanitizeValueForVega(config.x?.field);
  const yField = config.y?.field;
  const sanitizedYField = sanitizeValueForVega(yField);

  const defaultTooltipChannel = createDefaultTooltipEncoding(
    [config.x, config.y, config.color],
    data,
  );
  const multiValueTooltipChannel = createCartesianMultiValueTooltipChannel(
    { x: config.x, colorField, yField },
    data,
  );
  // Axis label layout depends on the channel the axis ends up on.
  spec.encoding = {
    x: {
      ...createPositionEncoding(config.x, data, horizontal ? "y" : "x"),
      bandPosition: 0,
    },
  };

  // Check if comparison mode is enabled
  const hasComparison = data.hasComparison;

  // Brushing is tied to the x channel, so it is disabled for horizontal charts.
  const isInteractive = !!config.isInteractive && !horizontal;

  const hoverRuleLayer = buildHoverRuleLayer({
    xField: sanitizedXField,
    domainValues: data.domainValues,
    isBarMark: true,
    defaultTooltip: defaultTooltipChannel,
    multiValueTooltipChannel,
    xSort: config.x?.sort,
    primaryColor: data.theme.primary,
    isDarkMode: data.isDarkMode,
    isInteractive,
    pivot: createVegaTransformPivotConfig(
      sanitizedXField,
      sanitizedYField,
      colorField,
      !!hasComparison,
      !!multiValueTooltipChannel?.length,
    ),
  });

  const barLayer: UnitSpec<Field> = {
    mark: { type: "bar", clip: true, width: { band: 0.9 } },
    encoding: {
      y: {
        ...createPositionEncoding(config.y, data, "y"),
        ...createStackOverride(config.y, colorField),
      },
      color: createColorEncoding(config.color, data),
      tooltip: defaultTooltipChannel,
    },
  };

  if (hasComparison && colorField) {
    // Comparison mode for stacked bars: use transforms with color dimension
    const transforms = createComparisonTransforms(
      config.x?.field,
      config.y?.field,
      colorField,
    );

    spec.transform = transforms;
    barLayer.encoding!.xOffset = createComparisonXOffsetEncoding();
    barLayer.encoding!.opacity = createComparisonOpacityEncoding(yField);
  } else if (hasComparison) {
    const transforms = createComparisonTransforms(
      config.x?.field,
      config.y?.field,
    );

    spec.transform = transforms;
    barLayer.encoding!.xOffset = createComparisonXOffsetEncoding();
    barLayer.encoding!.opacity = createComparisonOpacityEncoding(yField);
  }

  spec.layer = [hoverRuleLayer, barLayer];

  const result: VisualizationSpec = {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
    ...(isInteractive && sanitizedXField
      ? { usermeta: { brushTemporalField: sanitizedXField } }
      : {}),
  };

  return horizontal ? transposeCartesianSpec(result) : result;
}
