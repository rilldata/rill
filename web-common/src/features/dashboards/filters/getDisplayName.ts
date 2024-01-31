import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
} from "@rilldata/web-common/runtime-client";

export function getDimensionDisplayName(
  dimension: MetricsViewSpecDimensionV2 | undefined,
) {
  return (dimension?.label?.length ? dimension.label : dimension?.name) ?? "";
}

export function getMeasureDisplayName(
  measure: MetricsViewSpecMeasureV2 | undefined,
) {
  return (measure?.label?.length ? measure.label : measure?.expression) ?? "";
}
