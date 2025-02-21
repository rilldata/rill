import type { V1Expression } from "@rilldata/web-admin/client/gen/index.schemas";
import { getDimensionFilterWithSearch } from "../dashboards/dimension-table/dimension-table-utils";
import { mergeDimensionAndMeasureFilters } from "../dashboards/filters/measure-filters/measure-filter-utils";
import { sanitiseExpression } from "../dashboards/stores/filter-utils";
import type { DimensionThresholdFilter } from "../dashboards/stores/metrics-explorer-entity";

/**
 * Builds the where param for dimension table and TDD exports.
 *
 * If there's input in the search field, then all search results will be included in the export.
 * Otherwise, use the dashboard's current where filter.
 */
export function buildWhereParamForDimensionTableAndTDDExports(
  whereFilter: V1Expression,
  dimensionThresholdFilters: DimensionThresholdFilter[],
  dimensionName: string,
  searchText: string,
) {
  let dimensionFilter: V1Expression | undefined;
  if (searchText) {
    dimensionFilter = getDimensionFilterWithSearch(
      whereFilter,
      searchText,
      dimensionName,
    );
  } else {
    dimensionFilter = whereFilter;
  }

  const where = mergeDimensionAndMeasureFilters(
    dimensionFilter,
    dimensionThresholdFilters,
  );
  const sanitisedWhere = sanitiseExpression(where, undefined);
  return sanitisedWhere;
}
