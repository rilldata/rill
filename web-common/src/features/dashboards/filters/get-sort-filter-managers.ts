import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterManager.svelte.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterManager.svelte.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import type { JoinerFilterManager } from "@rilldata/web-common/features/dashboards/filters/JoinerFilterManager.svelte.ts";

/**
 * Managers in chip order: required filters first, then pinned ones, then the rest.
 * Sorted by name within each of those groups.
 */
export function getSortFilterManagers(
  joinerManager: JoinerFilterManager,
  yamlConfigProvider: YAMLConfigProvider,
) {
  const rank = (manager: DimensionFilterManager | MeasureFilterManager) => {
    if (yamlConfigProvider.requiredFilters[manager.name]) return 0;
    if (yamlConfigProvider.pinnedFilters[manager.name]) return 1;
    return 2;
  };

  const sortFunc = (
    a: DimensionFilterManager | MeasureFilterManager,
    b: DimensionFilterManager | MeasureFilterManager,
  ) => rank(a) - rank(b) || a.name.localeCompare(b.name);

  // Sorted on a copy: the managers keep param order, which is the order the expression and the
  // param it writes back are built in.
  const dimensions = [...joinerManager.managers.dimensionManagers].sort(
    sortFunc,
  );
  const measures = [...joinerManager.managers.measureManagers].sort(sortFunc);

  return { dimensions, measures };
}
