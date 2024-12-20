import {
  ResourceKind,
  useFilteredResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";

export function useDefaultMetrics(instanceId: string) {
  return useFilteredResources(instanceId, ResourceKind.MetricsView, (data) => {
    const validMetricsViews = data?.resources?.filter(
      (res) => !!res.metricsView?.state?.validSpec,
    );

    if (validMetricsViews && validMetricsViews?.length > 0) {
      const firstMetricsView =
        validMetricsViews[0].metricsView?.state?.validSpec;
      const firstMetricName = validMetricsViews[0].meta?.name?.name as string;
      const firstMeasure = firstMetricsView?.measures?.[0]?.name as string;
      const firstDimension =
        firstMetricsView?.dimensions?.[0]?.name ||
        (firstMetricsView?.dimensions?.[0]?.column as string);

      return {
        metricsView: firstMetricName,
        measure: firstMeasure,
        dimension: firstDimension,
      };
    }

    return null;
  });
}
