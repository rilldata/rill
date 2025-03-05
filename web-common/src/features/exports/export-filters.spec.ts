import { describe, expect, it } from "vitest";
import {
  createAndExpression,
  createInExpression,
  createLikeExpression,
} from "../dashboards/stores/filter-utils";
import { buildWhereParamForDimensionTableAndTDDExports } from "./export-filters";

describe("buildWhereParamForDimensionTableAndTDDExports", () => {
  const DIMENSION_NAME = "Customer";

  // Common test setup
  const createTestParams = (
    whereFilter = createAndExpression([]),
    searchText = "",
    measureFilters = [],
  ) => ({
    whereFilter,
    measureFilters,
    dimensionName: DIMENSION_NAME,
    searchText,
  });

  it("should return undefined when no filters or search text exist", () => {
    const params = createTestParams();

    const result = buildWhereParamForDimensionTableAndTDDExports(
      params.whereFilter,
      params.measureFilters,
      params.dimensionName,
      params.searchText,
    );

    expect(result).toBeUndefined();
  });

  it("should return whereFilter when only filters exist", () => {
    const whereFilter = createAndExpression([
      createInExpression(DIMENSION_NAME, ["Facebook"]),
    ]);
    const params = createTestParams(whereFilter);

    const result = buildWhereParamForDimensionTableAndTDDExports(
      params.whereFilter,
      params.measureFilters,
      params.dimensionName,
      params.searchText,
    );

    expect(result).toEqual(whereFilter);
  });

  it("should prioritize search text over filters when both exist", () => {
    const whereFilter = createAndExpression([
      createInExpression(DIMENSION_NAME, ["Facebook"]),
    ]);
    const searchText = "Face";
    const params = createTestParams(whereFilter, searchText);

    const result = buildWhereParamForDimensionTableAndTDDExports(
      params.whereFilter,
      params.measureFilters,
      params.dimensionName,
      params.searchText,
    );

    const expectedFilter = createAndExpression([
      createLikeExpression(DIMENSION_NAME, `%${searchText}%`),
    ]);
    expect(result).toEqual(expectedFilter);
  });
});
