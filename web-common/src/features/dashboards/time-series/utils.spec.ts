import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { describe, expect, it } from "vitest";
import { getFilterForComparedDimension, niceMeasureExtents } from "./utils";

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
  it("should remove filters for selected dimensions", () => {
    const dimensionName = "country";

    const filters = createAndExpression([
      createInExpression("company", ["zoom"]),
      createInExpression("country", ["IN"], true),
    ]);

    const result = getFilterForComparedDimension(dimensionName, filters);

    expect(result).toEqual(
      createAndExpression([createInExpression("company", ["zoom"])]),
    );
  });

  it("should not modify filters for unrelated dimensions", () => {
    const dimensionName = "country";

    const filters = createAndExpression([
      createInExpression("company", ["zoom"]),
      createInExpression("device", ["mobile"], true),
    ]);

    const result = getFilterForComparedDimension(dimensionName, filters);

    expect(result).toEqual(filters);
  });
});
