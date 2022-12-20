import type { MetricsViewDimension } from "@rilldata/web-common/runtime-client";
import { MetricsExplorerEntity } from "@rilldata/web-local/lib/application-state-stores/explorer-stores";

export function getDisplayName(dimension: MetricsViewDimension) {
  return dimension?.label?.length ? dimension?.label : dimension?.name;
}

export function hasDefinedTimeSeries(
  metricsExplorerStore: MetricsExplorerEntity
) {
  return !!metricsExplorerStore?.selectedTimeRange;
}
