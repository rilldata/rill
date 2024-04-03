import type { TopLevelSpec } from "vega-lite";

/*
  Read Data
  Come up with a heuristic to check if data has time fields and nominal fields
  Suggest chart type options based on data
  Limit to Bar, stack bar, line, area and stacked area charts
  Write a builder function to create the spec based on the data
  Add extents for TDD charts
  For rest add Template UI 
  */

export function buildVegaLiteSpec(
  chartType: string,
  timeFields: string[],
  quantitativeFields: string[],
  nominalFields: string[] = [],
): TopLevelSpec {
  const baseSpec: Partial<TopLevelSpec> = {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    description: `A ${chartType} chart.`,
    width: "container",
    data: { name: "table" },
  };

  // For now only support temporal data
  if (!timeFields.length) throw "No time fields found";

  const hasNominalFields = nominalFields.length > 0;

  // Encoding varies by chart type
  switch (chartType) {
    case "bar":
    case "stacked bar":
      baseSpec.mark = {
        type: "bar",
        width: { band: 0.75 },
        clip: true,
        bandPosition: 0,
      };
      baseSpec.encoding = {
        x: { field: timeFields[0], type: "temporal" },
        y: { field: quantitativeFields[0], type: "quantitative" },
        opacity: {
          condition: { param: "hover", empty: false, value: 1 },
          value: 0.8,
        },
        ...(hasNominalFields && {
          color: { field: nominalFields[0], type: "nominal", legend: null },
        }),
      };
      baseSpec.params = [
        {
          name: "hover",
          select: {
            type: "point",
            on: "pointerover",
          },
        },
      ];
      break;

    case "area":
      baseSpec.mark = { type: "area", clip: true };
      baseSpec.encoding = {
        x: { field: timeFields[0], type: "temporal" },
        y: { field: quantitativeFields[0], type: "quantitative" },
      };
      break;

    case "stacked area":
      baseSpec.layer = [
        {
          mark: { type: "area", clip: true },
          encoding: {
            x: { field: timeFields[0], type: "temporal" },
            y: {
              field: quantitativeFields[0],
              type: "quantitative",
              stack: "zero",
            },
            color: { field: nominalFields[0], type: "nominal", legend: null },
            opacity: {
              condition: { param: "hover", empty: false, value: 1 },
              value: 0.7,
            },
          },
          params: [
            {
              name: "hover",
              select: { type: "point", on: "pointerover" },
            },
          ],
        },
        {
          mark: { type: "line", strokeWidth: 1, clip: true },
          encoding: {
            x: { field: timeFields[0], type: "temporal" },
            y: {
              field: quantitativeFields[0],
              type: "quantitative",
              stack: "zero",
            },
            stroke: { field: nominalFields[0], type: "nominal", legend: null },
          },
        },
      ];
      break;

    case "line":
      baseSpec.mark = { type: "line", clip: true };
      baseSpec.encoding = {
        x: { field: timeFields[0], type: "temporal" },
        y: { field: quantitativeFields[0], type: "quantitative" },
        ...(hasNominalFields && {
          color: { field: nominalFields[0], type: "nominal" },
        }),
      };
      break;

    default:
      throw new Error(`Chart type "${chartType}" not supported.`);
  }

  return baseSpec as TopLevelSpec;
}
