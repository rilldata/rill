import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { LayerSpec } from "vega-lite/build/src/spec/layer";
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import {
  buildHoverRuleLayer,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../builder";
import type { ChartDataResult } from "../types";
import type { ComboChartSpec } from "./ComboChart";

function getColorForField(fieldName: string, config: ComboChartSpec): string {
  const colorMapping = config.color?.colorMapping;

  if (colorMapping) {
    const mapping = colorMapping?.find?.((m) => m.value === fieldName);
    if (mapping) {
      return mapping.color;
    }
  }

  return "#3524C7"; // fallback
}

export function generateVLComboChartSpec(
  config: ComboChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, undefined);
  const xField = sanitizeValueForVega(config.x?.field);

  const y1MarkType = config.y1?.mark || "bar";
  const y2MarkType = config.y2?.mark || "line";

  const defaultTooltipChannel = createDefaultTooltipEncoding(
    [config.x, config.y1, config.y2],
    data,
  );

  spec.height = "container";
  spec.encoding = { x: createPositionEncoding(config.x, data) };

  const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [];

  // Add hover rule layer
  layers.push(
    buildHoverRuleLayer({
      xField,
      domainValues: data.domainValues,
      isBarMark: y1MarkType === "bar" || y2MarkType === "bar",
      defaultTooltip: defaultTooltipChannel,
      xSort: config.x?.sort,
      primaryColor: data.theme.primary,
      xBand: config.x?.type === "temporal" ? 0.5 : undefined,
    }),
  );

  // Collect all data layers first
  const dataLayers: Array<UnitSpec<Field>> = [];

  // Add Y1 layer
  if (config.y1?.field) {
    const y1Color = getColorForField(config.y1.field, config);
    const y1Layer: UnitSpec<Field> = {
      mark: {
        type: y1MarkType,
        clip: true,
        ...(y1MarkType === "bar" && { width: { band: 0.9 } }),
        ...(y1MarkType === "line" && {
          point: true,
          strokeWidth: 2,
          interpolate: "monotone",
        }),
      },
      encoding: {
        y: {
          ...createPositionEncoding(config.y1, data),
        },
        color: { value: y1Color },
      },
    };
    dataLayers.push(y1Layer);
  }

  // Add Y2 layer
  if (config.y2?.field) {
    const y2Color = getColorForField(config.y2.field, config);
    const y2Layer: UnitSpec<Field> = {
      mark: {
        type: y2MarkType,
        clip: true,
        ...(y2MarkType === "bar" && { width: { band: 0.9 } }),
        ...(y2MarkType === "line" && {
          point: true,
          strokeWidth: 2,
          interpolate: "monotone",
        }),
      },
      encoding: {
        y: {
          ...createPositionEncoding(config.y2, data),
          axis: {
            ...createPositionEncoding(config.y2, data).axis,
            orient: "right",
          },
        },
        color: { value: y2Color },
      },
    };
    dataLayers.push(y2Layer);
  }

  // Sort layers so bar layers come first, then line layers (ensuring lines are on top)
  dataLayers.sort((unitA, unitB) => {
    const aMarkType =
      typeof unitA.mark === "string" ? unitA.mark : unitA.mark?.type || "bar";
    const bMarkType =
      typeof unitB.mark === "string" ? unitB.mark : unitB.mark?.type || "bar";

    // Bar layers should come before line layers
    if (aMarkType === "bar" && bMarkType === "line") return -1;
    if (aMarkType === "line" && bMarkType === "bar") return 1;
    return 0;
  });

  // Add sorted data layers to the main layers array
  layers.push(...dataLayers);

  spec.layer = layers;

  // Add dual axis resolution
  spec.resolve = {
    scale: { y: "independent" },
    axis: { y: "independent" },
  };

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
