import type { V1Expression } from "@rilldata/web-admin/client/gen/index.schemas";
import { getDimensionFilterWithSearch } from "../dashboards/dimension-table/dimension-table-utils";
import { sanitiseExpression } from "../dashboards/stores/filter-utils";

/**
 * If there's input in the search field, then all search results will be included in the export.
 * Otherwise, use the dashboard's current where filter.
 */
export function buildWhereParamForDimensionTableAndTDDExports(
  whereFilter: V1Expression | undefined,
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

  return sanitiseExpression(dimensionFilter, undefined);
}
