import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { ActiveValues } from "$lib/application-state-stores/explorer-stores";

// prepare the activeFilters to be sent to the server
export function prune(
  actives: ActiveValues,
  dimensions: Record<string, DimensionDefinitionEntity>
) {
  const filters: ActiveValues = {};
  for (const activeColumnId in actives) {
    if (!actives[activeColumnId].length) continue;
    filters[dimensions[activeColumnId].dimensionColumn] =
      actives[activeColumnId];
  }
  return filters;
}

export function getDisplayName(dimension) {
  return dimension?.labelSingle?.length
    ? dimension?.labelSingle
    : dimension?.dimensionColumn;
}
