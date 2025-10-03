import { CHART_CONFIG } from "@rilldata/web-common/features/components/charts/config";
import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
import { timeGrainToDuration } from "@rilldata/web-common/lib/time/grains";
import {
  V1TimeGrain,
  type V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import type { Color } from "chroma-js";
import merge from "deepmerge";
import type { Config } from "vega-lite";
import type {
  ChartDataResult,
  ChartDomainValues,
  ChartSpec,
  ChartType,
  ColorMapping,
  FieldConfig,
} from "./types";

export function isFieldConfig(field: unknown): field is FieldConfig {
  return (
    typeof field === "object" &&
    field !== null &&
    "type" in field &&
    "field" in field
  );
}

export function isMultiFieldConfig(field: unknown): field is FieldConfig {
  return isFieldConfig(field) && !!field.fields && field.fields.length > 0;
}

export function mergedVlConfig(
  userProvidedConfig: string | undefined,
  specConfig: Config | undefined,
): Config | undefined {
  if (!userProvidedConfig) return specConfig;

  const validSpecConfig = specConfig || {};
  let parsedConfig: Config;

  try {
    parsedConfig = JSON.parse(userProvidedConfig) as Config;
  } catch {
    console.warn("Invalid JSON config");
    return specConfig;
  }

  const replaceByClonedSource = (
    destinationArray: unknown[],
    sourceArray: unknown[],
  ) => sourceArray;

  return merge(validSpecConfig, parsedConfig, {
    arrayMerge: replaceByClonedSource,
  });
}

export interface FieldsByType {
  measures: string[];
  dimensions: string[];
  timeDimensions: string[];
}

export function getFieldsByType(spec: ChartSpec): FieldsByType {
  const measures = new Set<string>();
  const dimensions = new Set<string>();
  const timeDimensions = new Set<string>();

  // Recursively check all properties for FieldConfig objects
  const checkFields = (obj: unknown): void => {
    if (!obj || typeof obj !== "object") {
      return;
    }

    // Check if current object is a FieldConfig with type and field
    if (isFieldConfig(obj)) {
      const type = obj.type as string;
      const field = obj.field;
      const fields = obj.fields;

      switch (type) {
        case "quantitative":
          measures.add(field);
          if (fields) {
            fields.forEach((f) => measures.add(f));
          }
          break;
        case "nominal":
          dimensions.add(field);
          if (fields) {
            fields.forEach((f) => dimensions.add(f));
          }
          break;
        case "temporal":
          timeDimensions.add(field);
          if (fields) {
            fields.forEach((f) => timeDimensions.add(f));
          }
          break;
      }
      return;
    }

    Object.values(obj).forEach((value) => {
      if (typeof value === "object" && value !== null) {
        checkFields(value);
      }
    });
  };

  checkFields(spec);

  return {
    measures: Array.from(measures),
    dimensions: Array.from(dimensions),
    timeDimensions: Array.from(timeDimensions),
  };
}

export function adjustDataForTimeZone(
  data: V1MetricsViewAggregationResponseDataItem[] | undefined,
  timeFields: string[],
  timeGrain: V1TimeGrain,
  selectedTimezone: string,
) {
  if (!data) return data;

  return data.map((datum) => {
    // Create a shallow copy of the datum to avoid mutating the original
    const adjustedDatum = { ...datum };
    timeFields.forEach((timeField) => {
      adjustedDatum[timeField] = adjustOffsetForZone(
        datum[timeField] as string,
        selectedTimezone,
        timeGrainToDuration(timeGrain),
      );
    });
    return adjustedDatum;
  });
}

export function resolveColor(
  theme: { primary: Color; secondary: Color },
  color: string,
): string {
  if (color === "primary") {
    // Vega lite requires scale hsl colors to be comma separated
    const hslColor = theme.primary
      .css("hsl")
      .replace("deg", "")
      .replaceAll(" ", ", ");
    return hslColor;
  } else if (color === "secondary") {
    return theme.secondary.css("hsl").replace("deg", "").replaceAll(" ", ", ");
  }
  return color;
}

export function generateSpec(
  chartType: ChartType,
  rillChartSpec: ChartSpec,
  data: ChartDataResult,
) {
  if (data.isFetching || data.error) return {};
  return CHART_CONFIG[chartType]?.generateSpec(rillChartSpec, data);
}

export function isDomainStringArray(
  values: string[] | number[] | undefined,
): values is string[] {
  return values
    ? Array.isArray(values) &&
        values.every((value) => typeof value === "string")
    : false;
}

export function getColorForValues(
  colorValues: string[] | undefined,
  // if provided, use the colors for mentioned values
  overrideColorMapping: ColorMapping | undefined,
): ColorMapping | undefined {
  if (!colorValues || colorValues.length === 0) return undefined;

  const colorMapping = colorValues.map((value, index) => {
    const overrideColor = overrideColorMapping?.find(
      (mapping) => mapping.value === value,
    );
    return {
      value,
      color:
        overrideColor?.color ||
        COMPARIONS_COLORS[index % COMPARIONS_COLORS.length],
    };
  });

  return colorMapping;
}

export function getColorMappingForChart(
  chartSpec: ChartSpec,
  domainValues: ChartDomainValues | undefined,
): ColorMapping | undefined {
  if (!("color" in chartSpec) || !domainValues) return undefined;
  const colorField = chartSpec.color;

  let colorMapping: ColorMapping | undefined;
  if (typeof colorField === "object") {
    const fieldKey = colorField.field;
    const colorValues = domainValues[fieldKey];
    if (isDomainStringArray(colorValues)) {
      colorMapping = getColorForValues(colorValues, colorField.colorMapping);
    }
  }

  return colorMapping;
}
