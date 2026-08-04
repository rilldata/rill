import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";

export function getMissingRequiredFilters(
  expressionFilterManager: ExpressionFilterManager,
  yamlConfigProvider: YAMLConfigProvider,
) {
  return Object.keys(yamlConfigProvider.requiredFilters).filter(
    (filterName) => {
      return !expressionFilterManager.filterManagersMap[filterName]?.expr;
    },
  );
}
