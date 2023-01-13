import type {
  MetricsViewDimension,
  V1MetricsView,
} from "@rilldata/web-common/runtime-client";

export function getDisplayName(dimension: MetricsViewDimension) {
  return dimension?.label?.length ? dimension?.label : dimension?.name;
}

export function hasDefinedTimeSeries(metricsView: V1MetricsView) {
  return !!metricsView?.timeDimension;
}
