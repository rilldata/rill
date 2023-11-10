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
  it("should return filter with dimension added if no existing filter", () => {
    const dimensionName = "country";
    const filters = { include: [], exclude: [] };
    const topListValues = ["US", "IN", "CN"];

    const result = getFilterForComparedDimension(
      dimensionName,
      filters,
      topListValues,
      "chart"
    );

    expect(result.updatedFilter).toEqual({
      include: [{ name: "country", in: [] }],
      exclude: [],
    });

    expect(result.includedValues).toEqual(["US", "IN", "CN"]);
  });

  it("should exclude values from top list based on existing exclude filter", () => {
    const dimensionName = "country";
    const filters = {
      include: [],
      exclude: [{ name: "country", in: ["CN"] }],
    };
    const topListValues = ["US", "IN", "CN"];

    const result = getFilterForComparedDimension(
      dimensionName,
      filters,
      topListValues,
      "chart"
    );

    expect(result.updatedFilter).toEqual({
      include: [{ name: "country", in: [] }],
      exclude: [{ name: "country", in: ["CN"] }],
    });

    expect(result.includedValues).toEqual(["US", "IN"]);
  });

  it("should slice top list values to max of 3 for chart", () => {
    const dimensionName = "country";
    const filters = { include: [], exclude: [] };
    const topListValues = ["US", "IN", "CN", "UK", "FR"];

    const result = getFilterForComparedDimension(
      dimensionName,
      filters,
      topListValues,
      "chart"
    );

    expect(result.includedValues).toHaveLength(3);
  });

  it("should slice top list values to max of 250 for table", () => {
    const dimensionName = "country";
    const filters = { include: [], exclude: [] };
    const topListValues = new Array(300)
      .fill(null)
      .map((_, i) => `Country ${i}`);

    const result = getFilterForComparedDimension(
      dimensionName,
      filters,
      topListValues,
      "table"
    );

    expect(result.includedValues).toHaveLength(250);
  });

  it("should not modify filters for unrelated dimensions", () => {
    const dimensionName = "country";

    const filters = {
      include: [{ name: "company", in: ["zoom"] }],
      exclude: [{ name: "device", in: ["mobile"] }],
    };

    const topListValues = ["US", "IN", "CN"];

    const result = getFilterForComparedDimension(
      dimensionName,
      filters,
      topListValues,
      "chart"
    );

    expect(result.updatedFilter).toEqual({
      include: [
        { name: "company", in: ["zoom"] },
        { name: "country", in: [] },
      ],
      exclude: [{ name: "device", in: ["mobile"] }],
    });

    expect(result.includedValues).toEqual(["US", "IN", "CN"]);
  });
});
