import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { UnitSpec } from "vega-lite/build/src/spec";
import type { Transform } from "vega-lite/build/src/transform";
import {
  createConfig,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../builder";
import type { ChartDataResult } from "../types";
import { resolveCSSVariable } from "../util";
import type { FunnelChartSpec } from "./FunnelChartProvider";
import {
  createFunnelSortEncoding,
  getFormatType,
  getMultiMeasures,
} from "./util";

function createModeTransforms(
  mode: string | undefined,
  measureField: string | undefined,
  isMultiMeasure: boolean,
): Transform[] {
  if (mode === "order") {
    return [
      {
        window: [{ op: "row_number", as: "funnel_rank" }],
        sort: [
          {
            field: isMultiMeasure ? "value" : measureField!,
            order: "descending",
          },
        ],
      },
      {
        calculate: `pow(0.85, datum.funnel_rank - 1)`,
        as: "funnel_width",
      },
    ];
  } else {
    return [
      {
        calculate: isMultiMeasure
          ? `datum.value`
          : `datum['${sanitizeValueForVega(measureField!)}']`,
        as: "funnel_width",
      },
    ];
  }
}

function createPercentageTransforms(
  measureField: string | undefined,
  funnelSort: string[] | null,
  stageField: string | undefined,
  isMultiMeasure: boolean,
): Transform[] {
  if (isMultiMeasure) {
    return [
      {
        window: [
          {
            op: "first_value",
            field: "value",
            as: "reference_value",
          },
        ],
      },
      {
        calculate: `round((datum.value / datum.reference_value) * 100) + '%'`,
        as: "percentage",
      },
    ];
  } else {
    const transforms: Transform[] = [];

    if (Array.isArray(funnelSort) && funnelSort.length > 0 && stageField) {
      // Use joinaggregate to create a reference value field
      const firstStageInSort = funnelSort[0];
      transforms.push(
        {
          // Mark rows that match the first stage in custom sort
          calculate: `datum['${sanitizeValueForVega(stageField)}'] === '${sanitizeValueForVega(firstStageInSort)}' ? datum['${sanitizeValueForVega(measureField!)}'] : 0`,
          as: "is_reference_stage",
        },
        {
          // Use joinaggregate to get the maximum value where is_reference_stage > 0
          // This gives us the measure value for the first stage in custom sort
          joinaggregate: [
            {
              op: "max",
              field: "is_reference_stage",
              as: "reference_value",
            },
          ],
        },
      );
    } else {
      // For non-custom sort, use the first value in data order
      transforms.push({
        window: [
          {
            op: "first_value",
            field: measureField!,
            as: "reference_value",
          },
        ],
      });
    }

    transforms.push({
      calculate: `round((datum['${sanitizeValueForVega(measureField!)}'] / datum.reference_value) * 100) + '%'`,
      as: "percentage",
    });

    return transforms;
  }
}

export function generateVLFunnelChartSpec(
  config: FunnelChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createMultiLayerBaseSpec();
  spec.height = 500;

  const textColor = data.isDarkMode ? "#eeeeee" : "#353535";

  const isMultiMeasure = config.breakdownMode === "measures";

  const measureDisplayNames: Record<string, string> = {};
  if (isMultiMeasure && config.measure?.field) {
    const measures = getMultiMeasures(config.measure);
    measures.forEach((measure) => {
      measureDisplayNames[measure] =
        data.fields[measure]?.displayName || measure;
    });
  }
  let colorField: string | undefined;
  let colorType: string = "nominal";

  if (isMultiMeasure) {
    if (config.color === "name") {
      colorField = "Measure";
      colorType = "nominal";
    } else if (config.color === "value") {
      colorField = "value";
      colorType = "quantitative";
    }
  } else {
    // In dimension mode, color by stage or measure
    colorField =
      config.color === "measure" ? config.measure?.field : config.stage?.field;
    colorType =
      config.color === "measure"
        ? config.measure?.type || "quantitative"
        : "nominal";
  }

  const vegaConfig = createConfig(config);

  const yEncoding = isMultiMeasure
    ? { field: "Measure", type: "nominal" as const }
    : createPositionEncoding(config.stage, data);

  const tooltip = isMultiMeasure
    ? [
        {
          field: "measure_label",
          type: "nominal" as const,
          title: "Measure",
        },
        { field: "value", title: "Value", type: "quantitative" as const },
      ]
    : createDefaultTooltipEncoding([config.stage, config.measure], data);

  const funnelSort = isMultiMeasure
    ? getMultiMeasures(config.measure)
    : createFunnelSortEncoding(config.stage?.sort);

  // Add transforms
  const transforms: Transform[] = [];

  if (isMultiMeasure) {
    const measures = getMultiMeasures(config.measure);

    // Transform data for multi-measure funnel
    transforms.push(
      {
        fold: measures,
        as: ["Measure", "value"],
      },
      {
        calculate:
          Object.entries(measureDisplayNames)
            .map(([key, value]) => `datum.Measure === '${key}' ? '${value}' : `)
            .join("") + "datum.Measure",
        as: "measure_label",
      },
    );
  }

  if (config.measure?.field || isMultiMeasure) {
    const modeTransforms = createModeTransforms(
      config.mode,
      config.measure?.field,
      isMultiMeasure,
    );

    const percentageTransforms = createPercentageTransforms(
      config.measure?.field,
      Array.isArray(funnelSort) ? funnelSort : null,
      config.stage?.field,
      isMultiMeasure,
    );

    transforms.push(...modeTransforms, ...percentageTransforms);
  }

  if (transforms.length > 0) {
    spec.transform = transforms;
  }

  spec.encoding = {
    y: {
      ...yEncoding,
      sort: funnelSort,
      axis: {
        labels: false,
        title: null,
        ticks: false,
        domain: false,
      },
    },
    tooltip,
  };

  // Main bar layer
  const barLayer: UnitSpec<Field> = {
    mark: "bar",
    encoding: {
      x: {
        field: "funnel_width",
        type: "quantitative",
        stack: "center",
        axis: {
          labels: false,
          title: null,
          ticks: false,
          domain: false,
        },
      },
      color: {
        field: colorField,
        type: (colorType === "value" ? "nominal" : colorType) as
          | "quantitative"
          | "ordinal"
          | "nominal"
          | "temporal",
        legend: null,
      },
    },
  };

  const percentageTextLayer: UnitSpec<Field> = {
    mark: {
      type: "text",
      dx: {
        expr: `scale('x', datum['funnel_width']) < 10 ? 20 : 10`,
      },
      align: "left",
      fontWeight: 600,
      color: textColor,
    },
    encoding: {
      x: {
        field: "funnel_width",
        type: "quantitative",
        stack: "center",
      },
      text: {
        field: "percentage",
        type: "nominal",
      },
    },
  };

  // Resolve CSS variables for canvas rendering
  const darkTextColor = resolveCSSVariable(
    "var(--color-gray-900)",
    data.isDarkMode,
  );
  const lightTextColor = resolveCSSVariable(
    "var(--color-gray-50)",
    data.isDarkMode,
  );

  const valueTextLayer: UnitSpec<Field> = {
    mark: {
      type: "text",
      fontWeight: 600,
      dx: {
        expr: `-(scale('x', datum['funnel_width']) / 2)`,
      },
      color: {
        // Use theme-aware colors: dark text on light backgrounds, light text on dark backgrounds
        expr: `luminance ( scale ( 'color', datum['${sanitizeValueForVega(colorField ?? "")}'] ) ) > 0.45 ? '${darkTextColor}' : '${lightTextColor}'`,
      },
    },
    encoding: {
      x: {
        field: "funnel_width",
        type: "quantitative",
        stack: "center",
      },
      text: {
        field: isMultiMeasure ? "value" : config.measure?.field,
        type: isMultiMeasure
          ? "quantitative"
          : config.measure?.type === "value"
            ? "nominal"
            : config.measure?.type || "quantitative",
        formatType: getFormatType(config.measure, isMultiMeasure),
      },
    },
  };

  const labelTextLayer: UnitSpec<Field> = {
    mark: {
      type: "text",
      dx: {
        expr: `scale('x', datum['funnel_width']) < 10 ? -20 : -(scale('x', datum['funnel_width'])) - 10`,
      },
      align: "right",
      limit: 200,
      color: textColor,
    },
    encoding: {
      x: {
        field: "funnel_width",
        type: "quantitative",
        stack: "center",
      },
      text: isMultiMeasure
        ? {
            field: "measure_label",
            type: "nominal",
          }
        : {
            field: config.stage?.field,
            type: "nominal",
          },
    },
  };

  spec.layer = [barLayer, percentageTextLayer, valueTextLayer, labelTextLayer];

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
