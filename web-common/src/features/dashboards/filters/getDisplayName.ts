import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
} from "@rilldata/web-common/runtime-client";

export function getDimensionDisplayName(
  dimension: MetricsViewSpecDimensionV2 | undefined,
) {
  return (
    (dimension?.displayName?.length
      ? dimension.displayName
      : dimension?.name) ?? ""
  );
}

export function getMeasureDisplayName(
  measure: MetricsViewSpecMeasureV2 | undefined,
) {
  return (
    (measure?.displayName?.length
      ? measure.displayName
      : measure?.expression) ?? ""
  );
}
