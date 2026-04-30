import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/types_unstable/channeldef.js";
import type { UnitSpec } from "vega-lite/types_unstable/spec/unit.js";
import type { Transform } from "vega-lite/types_unstable/transform.js";
import {
  createConfig,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../builder";
import type { ChartDataResult } from "../types";
import { resolveCSSVariable } from "../util";
import type { FunnelChartSpec, FunnelPercentMode } from "./FunnelChartProvider";
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

// Format a numeric percentage (0-100) at one decimal, then drop a trailing `.0`
// so 32.1% stays "32.1%" but 32.0% becomes "32%". Keeps round numbers clean
// while letting neighboring values stay visually distinct.
function formatPercentExpr(valueExpr: string): string {
  return (
    `(!isValid(${valueExpr}) ? '' : ` +
    `replace(format(${valueExpr}, '.1f'), /\\.0$/, '') + '%')`
  );
}

function createPercentageTransforms(
  measureField: string | undefined,
  funnelSort: string[] | null,
  stageField: string | undefined,
  isMultiMeasure: boolean,
  percentMode: FunnelPercentMode | undefined,
): Transform[] {
  const valueExpr = isMultiMeasure
    ? "datum.value"
    : `datum['${sanitizeValueForVega(measureField!)}']`;
  const transforms: Transform[] = [];

  if (isMultiMeasure) {
    // Multi-measure: rows arrive in fold order, which matches the user-configured
    // measure order. Use that as the funnel order for both reference computations.
    transforms.push({
      window: [
        { op: "first_value", field: "value", as: "top_value" },
        { op: "lag", field: "value", as: "prev_value" },
      ],
    });
  } else if (Array.isArray(funnelSort) && funnelSort.length > 0 && stageField) {
    // Custom sort: explicit stage order. Tag each row with its position in the
    // sort array so we can derive both top and previous values via window ops.
    const sanitizedStage = sanitizeValueForVega(stageField);
    const positionExpr =
      funnelSort
        .map(
          (stage, idx) =>
            `datum['${sanitizedStage}'] === '${sanitizeValueForVega(stage)}' ? ${idx} : `,
        )
        .join("") + `${funnelSort.length}`;

    transforms.push(
      { calculate: positionExpr, as: "funnel_position" },
      {
        window: [
          { op: "first_value", field: measureField!, as: "top_value" },
          { op: "lag", field: measureField!, as: "prev_value" },
        ],
        sort: [{ field: "funnel_position", order: "ascending" }],
      },
    );
  } else {
    // Default sort: data is sorted by measure (asc/desc) at query time, so first
    // row is the top of the funnel and lag returns the immediately prior stage.
    transforms.push({
      window: [
        { op: "first_value", field: measureField!, as: "top_value" },
        { op: "lag", field: measureField!, as: "prev_value" },
      ],
    });
  }

  transforms.push(
    {
      calculate: `(${valueExpr} / datum.top_value) * 100`,
      as: "pct_of_top_num",
    },
    {
      // Top stage has no previous; show 100% there.
      calculate: `!isValid(datum.prev_value) ? 100 : (${valueExpr} / datum.prev_value) * 100`,
      as: "pct_of_prev_num",
    },
    {
      calculate: formatPercentExpr("datum.pct_of_top_num"),
      as: "pct_of_top",
    },
    {
      calculate: formatPercentExpr("datum.pct_of_prev_num"),
      as: "pct_of_previous",
    },
    {
      calculate:
        percentMode === "previous"
          ? "datum.pct_of_previous"
          : "datum.pct_of_top",
      as: "percentage",
    },
  );

  return transforms;
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

  const percentTooltipFields = [
    {
      field: "pct_of_top",
      title: "% of top",
      type: "nominal" as const,
    },
    {
      field: "pct_of_previous",
      title: "% of previous",
      type: "nominal" as const,
    },
  ];

  const tooltip = isMultiMeasure
    ? [
        {
          field: "measure_label",
          type: "nominal" as const,
          title: "Measure",
        },
        { field: "value", title: "Value", type: "quantitative" as const },
        ...percentTooltipFields,
      ]
    : [
        ...createDefaultTooltipEncoding([config.stage, config.measure], data),
        ...percentTooltipFields,
      ];

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
      config.percentMode,
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
      fontSize: 14,
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
