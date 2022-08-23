import type { MetricViewRequestFilter } from "$common/rill-developer-service/MetricViewActions";

export function isFiltered(filters: MetricViewRequestFilter): boolean {
  if (!filters) return false;
  return filters.include.length > 0 || filters.exclude.length > 0;
}
