import type { MetricsViewSpecDimensionV2 } from "@rilldata/web-common/runtime-client";

export function getDisplayName(
  dimension: MetricsViewSpecDimensionV2 | undefined
): string {
  return (dimension?.label?.length ? dimension?.label : dimension?.name) ?? "";
}
