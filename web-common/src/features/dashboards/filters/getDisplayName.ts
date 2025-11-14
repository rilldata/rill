import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
  V1Resource,
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

export function getExploreDisplayName(exploreResource: V1Resource | undefined) {
  const name = exploreResource?.meta?.name?.name;
  const spec = exploreResource?.explore?.state?.validSpec;
  return spec?.displayName || name || "";
}
