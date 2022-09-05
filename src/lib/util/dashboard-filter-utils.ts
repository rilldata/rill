import type { MetricsViewDimensionValues } from "$common/rill-developer-service/MetricsViewActions";
import { removeIfExists } from "$common/utils/arrayUtils";

export function removeFilterIfExists(
  dimensionId: string,
  dimensionValue: string,
  filters: MetricsViewDimensionValues
) {
  const dimensionEntryIndex = filters.findIndex(
    (filter) => filter.name === dimensionId
  );
  if (dimensionEntryIndex === -1) return false;

  if (
    !removeIfExists(
      filters[dimensionEntryIndex].values,
      (value) => value === dimensionValue
    )
  )
    return false;

  if (filters[dimensionEntryIndex].values.length === 0) {
    filters.splice(dimensionEntryIndex, 1);
  }

  return true;
}
