import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import type {
  ChartSortDirection,
  FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";

export function createFunnelSortEncoding(sort: ChartSortDirection | undefined) {
  if (sort && Array.isArray(sort)) {
    return sort;
  }
  return null;
}

export function getFormatType(
  measure: FieldConfig | undefined,
  isMultiMeasure: boolean,
) {
  if (isMultiMeasure) {
    return "humanize";
  } else if (
    measure?.type === "quantitative" &&
    measure?.field &&
    !isMultiMeasure
  ) {
    return sanitizeFieldName(measure.field);
  }
}

export function getMultiMeasures(measure: FieldConfig | undefined): string[] {
  if (measure?.fields?.length) {
    return measure.fields;
  } else if (measure?.field) {
    return [measure.field];
  }
  return [];
}
