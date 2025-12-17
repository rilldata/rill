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
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import type { CartesianChartSpec } from "../CartesianChartProvider";
import { createVegaTransformPivotConfig } from "../util";

function createTargetLineLayers(
  config: CartesianChartSpec,
  data: ChartDataResult,
  xField: string,
  yField: string,
): Array<LayerSpec<Field> | UnitSpec<Field>> {
  if (!data.targets || data.targets.length === 0) {
    return [];
  }

  const targetLayers: Array<LayerSpec<Field> | UnitSpec<Field>> = [];
  
  // Get the measure name from y field
  const measureName = config.y?.field;
  if (!measureName) return [];

  // Get the actual time dimension field name (unsanitized)
  const timeFieldName = config.x?.field;
  if (!timeFieldName) return [];

  // Find targets for this measure
  const measureTargets = data.targets.filter(
    (t) => t.measure === measureName,
  );

  for (const target of measureTargets) {
    if (!target.values || target.values.length === 0) continue;

    // Transform target values to chart data format
    // Target values have: time (or time dimension name), value, target, target_name
    // The time field name in target values matches the time dimension name in the query
    const targetName = target.target?.targetName || target.target?.name || "Target";
    
    // Create a target line layer
    // Use dashed line style for targets
    const targetLayer: UnitSpec<Field> = {
      data: {
        values: target.values.map((v) => {
          // Use the time field from target values (which matches the time dimension name)
          // Fallback to 'time' if the field doesn't exist
          const timeValue = v[timeFieldName] ?? v.time;
          return {
            [xField]: timeValue,
            [yField]: v.value,
            target_name: targetName,
          };
        }),
      },
      mark: {
        type: "line",
        strokeDash: [5, 5],
        opacity: 0.7,
        stroke: data.theme.primary.css("hsl").replace("deg", "").replaceAll(" ", ", "),
      },
      encoding: {
        x: createPositionEncoding(config.x, data),
        y: {
          field: sanitizeValueForVega(yField),
          type: "quantitative",
          title: yField,
        },
      },
    };

    targetLayers.push(targetLayer);
  }

  return targetLayers;
}

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

  const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [
    hoverRuleLayer,
    lineLayer,
  ];

  // Add target lines if available
  if (data.targets && config.x?.type === "temporal") {
    const targetLayers = createTargetLineLayers(config, data, xField, yField);
    layers.push(...targetLayers);
  }

  spec.layer = layers;

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
