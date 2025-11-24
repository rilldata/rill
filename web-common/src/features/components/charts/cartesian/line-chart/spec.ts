import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
import {
  buildHoverPointOverlay,
  buildHoverRuleLayer,
  createCartesianMultiValueTooltipChannel,
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "@rilldata/web-common/features/components/charts/builder";
import {
  ColorWithComparisonField,
  createComparisonOpacityEncoding,
  createComparisonTransforms,
  MeasureKeyField,
} from "@rilldata/web-common/features/components/charts/comparison-builder";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { LayerSpec } from "vega-lite/build/src/spec/layer";
import type { CartesianChartSpec } from "../CartesianChartProvider";
import { createVegaTransformPivotConfig } from "../util";

export function generateVLLineChartSpec(
  config: CartesianChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, config.color);

  const colorField =
    typeof config.color === "object" ? config.color.field : undefined;
  const xField = sanitizeValueForVega(config.x?.field);
  const yField = sanitizeValueForVega(config.y?.field);

  const defaultTooltipChannel = createDefaultTooltipEncoding(
    [config.x, config.y, config.color],
    data,
  );
  const multiValueTooltipChannel = createCartesianMultiValueTooltipChannel(
    { x: config.x, colorField, yField },
    data,
  );

  spec.encoding = { x: createPositionEncoding(config.x, data) };

  // Check if comparison mode is enabled
  const hasComparison = data.hasComparison;

  const hoverRuleLayer = buildHoverRuleLayer({
    xField,
    domainValues: data.domainValues,
    defaultTooltip: defaultTooltipChannel,
    multiValueTooltipChannel,
    xSort: config.x?.sort,
    primaryColor: data.theme.primary,
    isDarkMode: data.isDarkMode,
    pivot: createVegaTransformPivotConfig(
      xField,
      yField,
      colorField,
      !!hasComparison,
      !!multiValueTooltipChannel?.length,
    ),
  });

  const lineLayer: LayerSpec<Field> = {
    encoding: {
      y: createPositionEncoding(config.y, data),
      color: createColorEncoding(config.color, data),
    },
    layer: [{ mark: "line" }, buildHoverPointOverlay()],
  };

  if (hasComparison && colorField) {
    // Comparison mode for lines with color dimension: use transforms
    const transforms = createComparisonTransforms(
      config.x?.field,
      config.y?.field,
      colorField,
    );

    spec.transform = transforms;

    // Use detail encoding to separate lines by color_with_comparison
    // while keeping the original color for legend and coloring
    lineLayer.encoding!.detail = {
      field: ColorWithComparisonField,
      type: "nominal",
    };
    lineLayer.encoding!.opacity = createComparisonOpacityEncoding(yField);
  } else if (hasComparison) {
    const transforms = createComparisonTransforms(
      config.x?.field,
      config.y?.field,
    );

    spec.transform = transforms;

    // Use detail encoding to separate current and comparison lines
    lineLayer.encoding!.detail = {
      field: MeasureKeyField,
      type: "nominal",
    };
    lineLayer.encoding!.opacity = createComparisonOpacityEncoding(yField);
  }

  spec.layer = [hoverRuleLayer, lineLayer];

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
