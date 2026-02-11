import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import type {
  ChartDataResult,
  ColorMapping,
} from "@rilldata/web-common/features/components/charts/types";
import { resolveCSSVariable } from "@rilldata/web-common/features/components/charts/util";
import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import type { VisualizationSpec } from "svelte-vega";
import type { ColorDef, Field } from "vega-lite/build/src/channeldef";
import type { LayerSpec } from "vega-lite/build/src/spec/layer";
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import {
  buildHoverRuleLayer,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../builder";
import type { ComboChartSpec } from "./ComboChartProvider";

function getColorForField(
  encoding: "y1" | "y2",
  label: string,
  config: ComboChartSpec,
): string {
  const colorMapping = config.color?.colorMapping;

  if (colorMapping) {
    const mapping = colorMapping?.find?.((m) => m.value === label);
    if (mapping) {
      return mapping.color;
    }
  }

  // Use qualitative palette colors for the two measures
  if (encoding === "y1") return COMPARIONS_COLORS[0];
  if (encoding === "y2") return COMPARIONS_COLORS[1];

  // Fallback to qualitative palette color 3
  return COMPARIONS_COLORS[2];
}

export function generateVLComboChartSpec(
  config: ComboChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const measureField = "Measure";
  const valueField = "value";

  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, config.color);
  const xField = sanitizeValueForVega(config.x?.field);

  const y1MarkType = config.y1?.mark || "bar";
  const y2MarkType = config.y2?.mark || "line";

  const defaultTooltipChannel = createDefaultTooltipEncoding(
    [config.x, config.y1, config.y2],
    data,
  );

  const measures: string[] = [];
  const measureDisplayNames: Record<string, string> = {};
  const colorMapping: ColorMapping = [];

  if (config.y1?.field) {
    measures.push(config.y1.field);
    measureDisplayNames[config.y1.field] =
      data.fields[config.y1.field]?.displayName || config.y1.field;
    colorMapping.push({
      value: config.y1.field,
      color: getColorForField(
        "y1",
        measureDisplayNames[config.y1.field],
        config,
      ),
    });
  }
  if (config.y2?.field) {
    measures.push(config.y2.field);
    measureDisplayNames[config.y2.field] =
      data.fields[config.y2.field]?.displayName || config.y2.field;
    colorMapping.push({
      value: config.y2.field,
      color: getColorForField(
        "y2",
        measureDisplayNames[config.y2.field],
        config,
      ),
    });
  }

  // Transform data to long format for legend
  spec.transform = [
    {
      fold: measures,
      as: [measureField, valueField],
    },
  ];

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
      isDarkMode: data.isDarkMode,
      xBand: config.x?.type === "temporal" ? 0.5 : undefined,
    }),
  );

  const legend = {
    labelExpr:
      Object.entries(measureDisplayNames)
        .map(([key, value]) => `datum.value === '${key}' ? '${value}' : `)
        .join("") + "datum.value",
  };

  const baseColorEncoding: ColorDef<Field> = {
    field: measureField,
    type: "nominal",
    // Only apply legend if orientation is not "none"
    ...(config.color?.legendOrientation !== "none" && { legend }),
    scale: {
      domain: colorMapping.map((m) => m.value),
      // Resolve CSS variables for canvas rendering
      range: colorMapping.map((m) => resolveCSSVariable(m.color)),
      type: "ordinal",
    },
  };

  const dataLayers: Array<UnitSpec<Field>> = [];

  if (config.y1?.field) {
    const y1Layer: UnitSpec<Field> = {
      transform: [
        {
          filter: `datum['${measureField}'] === '${config.y1.field}'`,
        },
      ],
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
        ...(y1MarkType === "line" &&
          config.x?.type === "temporal" && {
            x: {
              ...createPositionEncoding(config.x, data),
              bandPosition: 0.5,
            },
          }),
        y: {
          ...createPositionEncoding(config.y1, data),
          field: valueField,
          axis: {
            ...createPositionEncoding(config.y1, data).axis,
            orient: "left",
          },
        },
        color: baseColorEncoding,
      },
    };
    dataLayers.push(y1Layer);
  }

  if (config.y2?.field) {
    const y2Layer: UnitSpec<Field> = {
      transform: [
        {
          filter: `datum['${measureField}'] === '${config.y2.field}'`,
        },
      ],
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
        ...(y2MarkType === "line" &&
          config.x?.type === "temporal" && {
            x: {
              ...createPositionEncoding(config.x, data),
              bandPosition: 0.5,
            },
          }),
        y: {
          ...createPositionEncoding(config.y2, data),
          field: valueField,
          axis: {
            ...createPositionEncoding(config.y2, data).axis,
            orient: "right",
          },
        },
        color: baseColorEncoding,
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
