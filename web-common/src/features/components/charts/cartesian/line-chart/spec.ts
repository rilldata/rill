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
import type { Field } from "vega-lite/types_unstable/channeldef.js";
import type { LayerSpec } from "vega-lite/types_unstable/spec/layer.js";
import type { CartesianChartSpec } from "../CartesianChartProvider";
import { toVerticalSpec } from "../orientation";
import { createVegaTransformPivotConfig } from "../util";

export function generateVLLineChartSpec(
  chartConfig: CartesianChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  // Line charts are always drawn with the dimension on x. The reconciler
  // rejects a measure on x for them; reading the fields by role here keeps
  // the builder consistent with the provider, which queries the same way.
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

  const xEncoding = createPositionEncoding(config.x, data, "x");
  xEncoding.scale = { ...(xEncoding.scale ?? {}), padding: 8 };
  spec.encoding = { x: xEncoding };

  // Check if comparison mode is enabled
  const hasComparison = data.hasComparison;

  const hoverRuleLayer = buildHoverRuleLayer({
    xField: sanitizedXField,
    domainValues: data.domainValues,
    defaultTooltip: defaultTooltipChannel,
    multiValueTooltipChannel,
    xSort: config.x?.sort,
    primaryColor: data.theme.primary,
    isDarkMode: data.isDarkMode,
    isInteractive: config.isInteractive,
    pivot: createVegaTransformPivotConfig(
      sanitizedXField,
      sanitizedYField,
      colorField,
      !!hasComparison,
      !!multiValueTooltipChannel?.length,
    ),
  });

  const lineLayer: LayerSpec<Field> = {
    encoding: {
      y: createPositionEncoding(config.y, data, "y"),
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
    ...(config.isInteractive && sanitizedXField
      ? { usermeta: { brushTemporalField: sanitizedXField } }
      : {}),
  };
}
