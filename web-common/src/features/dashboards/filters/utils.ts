import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";

export function getMissingRequiredFilters(
  expressionFilterManager: ExpressionFilterManager,
  metricsViewsProvider: MetricsViewsProvider,
  yamlConfigProvider: YAMLConfigProvider,
) {
  return Object.keys(yamlConfigProvider.requiredFilters).filter(
    (filterName) => {
      const isValidDimensionOrMeasure =
        filterName in metricsViewsProvider.dimensionSpecs ||
        filterName in metricsViewsProvider.measureSpecs;
      const hasNoFilter =
        !expressionFilterManager.filterManagersMap[filterName]?.expr;
      return isValidDimensionOrMeasure && hasNoFilter;
    },
  );
}
