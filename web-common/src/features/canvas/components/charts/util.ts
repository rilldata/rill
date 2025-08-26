import type { CartesianChartSpec } from "@rilldata/web-common/features/canvas/components/charts/cartesian-charts/CartesianChart";
import type { ComboChartSpec } from "@rilldata/web-common/features/canvas/components/charts/combo-charts/ComboChart";
import type { HeatmapChartSpec } from "@rilldata/web-common/features/canvas/components/charts/heatmap-charts/HeatmapChart";
import type { ColorMapping } from "@rilldata/web-common/features/canvas/inspector/types";
import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
import { timeGrainToDuration } from "@rilldata/web-common/lib/time/grains";
import {
  V1TimeGrain,
  type V1MetricsViewAggregationResponseDataItem,
  type V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import merge from "deepmerge";
import type { Config } from "vega-lite";
import { CHART_CONFIG, type ChartSpec } from "./";
import type {
  ChartDataResult,
  ChartDomainValues,
  ChartSortDirection,
  ChartType,
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
export function generateSpec(
  chartType: ChartType,
  rillChartSpec: ChartSpec,
  data: ChartDataResult,
) {
  if (data.isFetching || data.error) return {};
  return CHART_CONFIG[chartType]?.generateSpec(rillChartSpec, data);
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
  let measures: string[] = [];
  const dimensions: string[] = [];
  const timeDimensions: string[] = [];

  // Recursively check all properties for FieldConfig objects
  const checkFields = (obj: unknown): void => {
    if (!obj || typeof obj !== "object") {
      return;
    }

    // Check if current object is a FieldConfig with type and field
    if (isFieldConfig(obj)) {
      const type = obj.type as string;
      const field = obj.field;

      switch (type) {
        case "quantitative":
          measures.push(field);
          break;
        case "nominal":
          dimensions.push(field);
          break;
        case "temporal":
          timeDimensions.push(field);
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

  if ("measures" in spec && Array.isArray(spec.measures)) {
    measures = spec.measures;
  }
  return {
    measures,
    dimensions,
    timeDimensions,
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

/**
 * Converts a Vega-style sort configuration to Rill's aggregation sort format.
 */
export function vegaSortToAggregationSort(
  encoder: "x" | "y",
  config: CartesianChartSpec | HeatmapChartSpec | ComboChartSpec,
  defaultSort: ChartSortDirection,
): V1MetricsViewAggregationSort | undefined {
  const encoderConfig = config[encoder];

  if (!encoderConfig) {
    return undefined;
  }

  let sort = encoderConfig.sort;

  if (!sort || Array.isArray(sort)) {
    sort = defaultSort;
  }

  let field: string | undefined;
  let desc: boolean = false;

  switch (sort) {
    case "x":
    case "-x":
      field = config.x?.field;
      desc = sort === "-x";
      break;
    case "y":
    case "-y":
      if ("y" in config) {
        field = config.y?.field;
      } else if ("y1" in config) {
        field = config.y1?.field;
      }
      desc = sort === "-y";
      break;
    case "color":
    case "-color":
      field = isFieldConfig(config.color) ? config.color.field : undefined;
      desc = sort === "-color";
      break;
    default:
      return undefined;
  }

  if (!field) return undefined;

  return {
    name: field,
    desc,
  };
}

const allowedTimeDimensionDetailTypes = [
  "line_chart",
  "area_chart",
  "stacked_bar",
  "stacked_bar_normalized",
  "bar_chart",
];

export const CanvasChartTypeToTDDChartType = {
  line_chart: TDDChart.DEFAULT,
  area_chart: TDDChart.STACKED_AREA,
  stacked_bar: TDDChart.STACKED_BAR,
  stacked_bar_normalized: TDDChart.STACKED_BAR,
  bar_chart: TDDChart.GROUPED_BAR,
};

export function getLinkStateForTimeDimensionDetail(
  spec: ChartSpec,
  type: ChartType,
): {
  canLink: boolean;
  measureName?: string;
  dimensionName?: string;
} {
  if (!allowedTimeDimensionDetailTypes.includes(type))
    return {
      canLink: false,
    };

  const hasXAxis = "x" in spec;
  const hasYAxis = "y" in spec;
  if (!hasXAxis || !hasYAxis)
    return {
      canLink: false,
    };

  const xAxis = spec.x;
  const yAxis = spec.y;

  if (isFieldConfig(xAxis) && isFieldConfig(yAxis)) {
    const colorDimension = spec.color;
    if (isFieldConfig(colorDimension)) {
      return {
        canLink: xAxis.type === "temporal",
        measureName: yAxis.field,
        dimensionName: colorDimension.field,
      };
    }

    return {
      canLink: xAxis.type === "temporal",
      measureName: yAxis.field,
    };
  }
  return {
    canLink: false,
  };
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
    const fieldKey = colorField.field || "measures";
    colorMapping = getColorForValues(
      domainValues[fieldKey],
      colorField.colorMapping,
    );
  }

  return colorMapping;
}
