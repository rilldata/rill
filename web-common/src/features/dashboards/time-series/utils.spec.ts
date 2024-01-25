import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { describe, it, expect } from "vitest";
import { niceMeasureExtents, getFilterForComparedDimension } from "./utils";

describe("niceMeasureExtents", () => {
  it("should return [0, 1] if both values are 0", () => {
    expect(niceMeasureExtents([0, 0], 1)).toEqual([0, 1]);
  });
  it("should ensure that if the minimum is positive, it returns 0", () => {
    expect(niceMeasureExtents([1, 2], 1)).toEqual([0, 2]);
  });
  it("should ensure that if the maximum is negative, it returns 0", () => {
    expect(niceMeasureExtents([-5, -2], 1)).toEqual([-5, 0]);
  });
  it("should inflate the minimum if it is negative", () => {
    expect(niceMeasureExtents([-5, 0], 2)).toEqual([-10, 0]);
  });
  it("should inflate the maximum if it is positive", () => {
    expect(niceMeasureExtents([0, 5], 2)).toEqual([0, 10]);
  });
});

describe("getFilterForComparedDimension", () => {
  it("should slice top list values to max of 250 for table", () => {
    const dimensionName = "country";
    const topListValues = new Array(300)
      .fill(null)
      .map((_, i) => `Country ${i}`);

    const result = getFilterForComparedDimension(
      dimensionName,
      createAndExpression([]),
      topListValues,
    );

    expect(result.includedValues).toHaveLength(250);
  });

  it("should remove filters for selected dimensions", () => {
    const dimensionName = "country";

    const filters = createAndExpression([
      createInExpression("company", ["zoom"]),
      createInExpression("country", ["IN"], true),
    ]);

    const topListValues = ["US", "IN", "CN"];

    const result = getFilterForComparedDimension(
      dimensionName,
      filters,
      topListValues,
    );

    expect(result.updatedFilter).toEqual(
      createAndExpression([createInExpression("company", ["zoom"])]),
    );

    expect(result.includedValues).toEqual(["US", "IN", "CN"]);
  });

  it("should not modify filters for unrelated dimensions", () => {
    const dimensionName = "country";

    const filters = createAndExpression([
      createInExpression("company", ["zoom"]),
      createInExpression("device", ["mobile"], true),
    ]);

    const topListValues = ["US", "IN", "CN"];

    const result = getFilterForComparedDimension(
      dimensionName,
      filters,
      topListValues,
    );

    expect(result.updatedFilter).toEqual(filters);

    expect(result.includedValues).toEqual(["US", "IN", "CN"]);
  });
});
