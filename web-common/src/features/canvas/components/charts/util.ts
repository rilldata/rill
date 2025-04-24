import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { sanitizeValueForVega } from "@rilldata/web-common/features/templates/charts/utils";
import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
import { timeGrainToDuration } from "@rilldata/web-common/lib/time/grains";
import {
  V1TimeGrain,
  type V1Expression,
  type V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import merge from "deepmerge";
import type { Config } from "vega-lite";
import { CHART_CONFIG, type ChartSpec } from "./";
import type { ChartDataResult, ChartType, FieldConfig } from "./types";

export function generateSpec(
  chartType: ChartType,
  rillChartSpec: ChartSpec,
  data: ChartDataResult,
) {
  if (data.isFetching || data.error) return {};
  return CHART_CONFIG[chartType].generateSpec(rillChartSpec, data);
}

export function isChartLineLike(chartType: ChartType) {
  return chartType === "line_chart" || chartType === "area_chart";
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

export const timeGrainToVegaTimeUnitMap: Record<V1TimeGrain, string> = {
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: "yearmonthdatehoursminutesseconds",
  [V1TimeGrain.TIME_GRAIN_SECOND]: "yearmonthdatehoursminutesseconds",
  [V1TimeGrain.TIME_GRAIN_MINUTE]: "yearmonthdatehoursminutes",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "yearmonthdatehours",
  [V1TimeGrain.TIME_GRAIN_DAY]: "yearmonthdate",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "yearweek",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "yearmonth",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "yearquarter",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "year",
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "yearmonthdate",
};

export function sanitizeFieldName(fieldName: string) {
  const specialCharactersRemoved = sanitizeValueForVega(fieldName);
  const sanitizedFieldName = specialCharactersRemoved.replace(" ", "__");

  /**
   * Add a prefix to the beginning of the field
   * name to avoid variables starting with a special
   * character or number.
   */
  return `rill_${sanitizedFieldName}`;
}

export interface FieldsByType {
  measures: string[];
  dimensions: string[];
  timeDimensions: string[];
}

export function getFieldsByType(spec: ChartSpec): FieldsByType {
  const measures: string[] = [];
  const dimensions: string[] = [];
  const timeDimensions: string[] = [];

  // Recursively check all properties for FieldConfig objects
  const checkFields = (obj: unknown): void => {
    if (!obj || typeof obj !== "object") {
      return;
    }

    // Check if current object is a FieldConfig with type and field
    if ("type" in obj && "field" in obj && typeof obj.field === "string") {
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
  return {
    measures,
    dimensions,
    timeDimensions,
  };
}

export function getFilterWithNullHandling(
  where: V1Expression | undefined,
  fieldConfig: FieldConfig | undefined,
): V1Expression | undefined {
  if (!fieldConfig || fieldConfig.showNull || fieldConfig.type !== "nominal") {
    return where;
  }

  const excludeNullFilter = createInExpression(fieldConfig.field, [null], true);
  return mergeFilters(where, excludeNullFilter);
}

export function adjustDataForTimeZone(
  data: V1MetricsViewAggregationResponseDataItem[] | undefined,
  timeFields: string[],
  timeGrain: V1TimeGrain,
  selectedTimezone: string,
) {
  if (!data) return data;

  return data.map((datum) => {
    timeFields.forEach((timeField) => {
      datum[timeField] = adjustOffsetForZone(
        datum[timeField] as string,
        selectedTimezone,
        timeGrainToDuration(timeGrain),
      );
    });
    return datum;
  });
}
