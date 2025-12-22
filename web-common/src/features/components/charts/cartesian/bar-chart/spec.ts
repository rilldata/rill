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
} from "@rilldata/web-common/features/components/charts/builder";
import {
  ColorWithComparisonField,
  createComparisonOpacityEncoding,
  createComparisonTransforms,
  createComparisonXOffsetEncoding,
  SortOrderField,
} from "@rilldata/web-common/features/components/charts/comparison-builder";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import type { CartesianChartSpec } from "../CartesianChartProvider";
import { createVegaTransformPivotConfig } from "../util";

export function generateVLBarChartSpec(
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
    isBarMark: true,
    defaultTooltip: defaultTooltipChannel,
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
      y: createPositionEncoding(config.y, data),
      color: createColorEncoding(config.color, data),
      // Only add xOffset for color field when NOT in comparison mode
      // In comparison mode, we'll handle xOffset separately to include comparison grouping
      ...(config.color &&
      typeof config.color === "object" &&
      config.x &&
      !hasComparison
        ? {
            xOffset: {
              field: config.color.field,
              title:
                data.fields[config.color.field]?.displayName ||
                config.color.field,
            },
          }
        : {}),
    },
  };

  if (hasComparison && colorField) {
    // Comparison mode with color dimension: use transforms to create synthetic fields
    const transforms = createComparisonTransforms(
      config.x?.field,
      config.y?.field,
      colorField,
    );

    spec.transform = transforms;

    // Use the synthetic color_with_comparison field for xOffset to group by both color and period
    barLayer.encoding!.xOffset = {
      field: ColorWithComparisonField,
      sort: { field: SortOrderField },
    };
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

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
