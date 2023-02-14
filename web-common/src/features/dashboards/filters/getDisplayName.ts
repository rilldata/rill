import type { MetricsViewDimension } from "@rilldata/web-common/runtime-client";

export function getDisplayName(dimension: MetricsViewDimension) {
  return dimension?.label?.length ? dimension?.label : dimension?.name;
}
