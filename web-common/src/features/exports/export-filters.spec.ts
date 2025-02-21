import { describe, expect, it } from "vitest";
import {
  createAndExpression,
  createInExpression,
  createLikeExpression,
} from "../dashboards/stores/filter-utils";
import { buildWhereParamForDimensionTableAndTDDExports } from "./export-filters";

// Test cases:
// 1. No filters
// 2. Filters
// 3. Filters and search text
describe("buildWhereParamForDimensionTableAndTDDExports", () => {
  it("should be undefined if there are no filters and no search text", () => {
    const whereFilter = createAndExpression([]);
    const measureFilters = [];
    const dimensionName = "dummyDimension";
    const searchText = "";

    const whereParam = buildWhereParamForDimensionTableAndTDDExports(
      whereFilter,
      measureFilters,
      dimensionName,
      searchText,
    );
    expect(whereParam).toBeUndefined();
  });

  it("should be the `whereFilter` if there are filters and no search text", () => {
    const whereFilter = createAndExpression([
      createInExpression("Customer", ["Facebook"]),
    ]);
    const measureFilters = [];
    const dimensionName = "Customer";
    const searchText = "";

    const whereParam = buildWhereParamForDimensionTableAndTDDExports(
      whereFilter,
      measureFilters,
      dimensionName,
      searchText,
    );
    expect(whereParam).toEqual(whereFilter);
  });

  it("should use the search text if there is search text", () => {
    const whereFilter = createAndExpression([
      createInExpression("Customer", ["Facebook"]),
    ]);
    const measureFilters = [];
    const dimensionName = "Customer";
    const searchText = "Facebook";

    const whereParam = buildWhereParamForDimensionTableAndTDDExports(
      whereFilter,
      measureFilters,
      dimensionName,
      searchText,
    );

    expect(whereParam).toEqual(
      createAndExpression([createLikeExpression("Customer", "%Facebook%")]),
    );
  });
});
