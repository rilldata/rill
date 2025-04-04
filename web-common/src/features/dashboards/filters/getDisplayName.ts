import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";

export function getDimensionDisplayName(
  dimension: MetricsViewSpecDimension | undefined,
) {
  return (
    (dimension?.displayName?.length
      ? dimension.displayName
      : dimension?.name) ?? ""
  );
}

export function getMeasureDisplayName(
  measure: MetricsViewSpecMeasure | undefined,
) {
  return (
    (measure?.displayName?.length
      ? measure.displayName
      : measure?.expression) ?? ""
  );
}
