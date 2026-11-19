import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterManager.svelte.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterManager.svelte.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import type { JoinerFilterManager } from "@rilldata/web-common/features/dashboards/filters/JoinerFilterManager.svelte.ts";

/**
 * Managers in chip order.
 * Required filters that are not also pinned come first, in the order of `filters.required`.
 * Pinned filters follow in the order of `filters.pinned`, which places a filter that is both pinned and required.
 * The other filters come last, sorted by name.
 */
export function getSortFilterManagers(
  joinerManager: JoinerFilterManager,
  yamlConfigProvider: YAMLConfigProvider,
) {
  const { pinnedFilters, requiredFilters } = yamlConfigProvider;
  const positions = new Map<string, number>();
  requiredFilters
    .filter((name) => !pinnedFilters.includes(name))
    .concat(pinnedFilters)
    .forEach((name, index) => positions.set(name, index));
  // Filters that are neither pinned nor required share the last position.
  const position = (manager: DimensionFilterManager | MeasureFilterManager) =>
    positions.get(manager.name) ?? positions.size;

  const sortFunc = (
    a: DimensionFilterManager | MeasureFilterManager,
    b: DimensionFilterManager | MeasureFilterManager,
  ) => position(a) - position(b) || a.name.localeCompare(b.name);

  // Sorted on a copy: the managers keep param order, which is the order the expression and the
  // param it writes back are built in.
  const dimensions = [...joinerManager.managers.dimensionManagers].sort(
    sortFunc,
  );
  const measures = [...joinerManager.managers.measureManagers].sort(sortFunc);

  return { dimensions, measures };
}

/**
 * Keys for a keyed `{#each}` over filter managers, aligned with the given list.
 * Managers are rebuilt on every URL change, so keying by the manager remounts each chip;
 * the name is stable, but an advanced filter holds a manager per condition,
 * so repeats of a name get their occurrence appended.
 */
export function getFilterManagerKeys(managers: { name: string }[]): string[] {
  const occurrences = new Map<string, number>();
  return managers.map(({ name }) => {
    const occurrence = (occurrences.get(name) ?? 0) + 1;
    occurrences.set(name, occurrence);
    return occurrence === 1 ? name : `${name}#${occurrence}`;
  });
}
