import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { type Document, YAMLMap } from "yaml";

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

/**
 * Writes `filters.pinned` and `filters.required` to a dashboard's YAML.
 * An empty list removes its key, and the `filters` block is removed once nothing is left in it.
 * Other keys in the block are left as they are.
 *
 * `keyPath` locates the dashboard's properties in the document:
 * empty for a canvas or explore file, `["explore"]` for an explore defined inside a metrics view.
 */
export function setPinnedAndRequiredFilters(
  yaml: Document,
  pinned: string[],
  required: string[],
  keyPath: string[] = [],
) {
  const filtersPath = [...keyPath, "filters"];
  Object.entries({ pinned, required }).forEach(([key, names]) => {
    if (names.length > 0) {
      yaml.setIn([...filtersPath, key], names);
    } else if (yaml.hasIn([...filtersPath, key])) {
      yaml.deleteIn([...filtersPath, key]);
    }
  });

  const filters = yaml.getIn(filtersPath);
  if (filters instanceof YAMLMap && filters.items.length === 0) {
    yaml.deleteIn(filtersPath);
  }
}
