import { describe, expect, it } from "vitest";
import {
  computeOtherGrouping,
  getOtherTooltipData,
  OTHER_SLICE_LABEL,
} from "./other-grouping";

function makeData(
  values: Array<[string, number]>,
  colorField = "category",
  measureField = "value",
) {
  return values.map(([name, val]) => ({
    [colorField]: name,
    [measureField]: val,
  }));
}

describe("computeOtherGrouping", () => {
  it("returns all slices when 3 or fewer categories", () => {
    const data = makeData([
      ["A", 50],
      ["B", 30],
      ["C", 20],
    ]);
    const result = computeOtherGrouping(data, "value", "category", {});
    expect(result.hasOther).toBe(false);
    expect(result.visibleData).toHaveLength(3);
    expect(result.total).toBe(100);
  });

  it("groups long-tail distribution (top 3 = 80%): shows ~4-5 slices + Other", () => {
    const data = makeData([
      ["A", 400],
      ["B", 250],
      ["C", 150],
      ["D", 80],
      ["E", 50],
      ["F", 30],
      ["G", 25],
      ["H", 15],
    ]);
    const result = computeOtherGrouping(data, "value", "category", {});
    expect(result.hasOther).toBe(true);
    const otherSlice = result.visibleData.find(
      (d) => d.category === OTHER_SLICE_LABEL,
    );
    expect(otherSlice).toBeDefined();
    const namedSlices = result.visibleData.filter(
      (d) => d.category !== OTHER_SLICE_LABEL,
    );
    expect(namedSlices.length).toBeGreaterThanOrEqual(3);
    expect(namedSlices.length).toBeLessThanOrEqual(7);
  });

  it("groups even distribution (20 categories, ~5% each): shows 10 + Other", () => {
    const data = makeData(
      Array.from({ length: 20 }, (_, i) => [`Cat${i}`, 50]),
    );
    const result = computeOtherGrouping(data, "value", "category", {});
    expect(result.hasOther).toBe(true);
    const namedSlices = result.visibleData.filter(
      (d) => d.category !== OTHER_SLICE_LABEL,
    );
    expect(namedSlices.length).toBe(10);
  });

  it("respects explicit limit: 3", () => {
    const data = makeData(
      Array.from({ length: 10 }, (_, i) => [`Cat${i}`, 100 - i * 10]),
    );
    const result = computeOtherGrouping(data, "value", "category", {
      limit: 3,
    });
    expect(result.hasOther).toBe(true);
    const namedSlices = result.visibleData.filter(
      (d) => d.category !== OTHER_SLICE_LABEL,
    );
    expect(namedSlices.length).toBe(3);
  });

  it("respects showOther: false", () => {
    const data = makeData(
      Array.from({ length: 20 }, (_, i) => [`Cat${i}`, 50]),
    );
    const result = computeOtherGrouping(data, "value", "category", {
      showOther: false,
    });
    expect(result.hasOther).toBe(false);
    expect(result.visibleData).toHaveLength(20);
  });

  it("does not create Other when only 1 item would be grouped", () => {
    const data = makeData([
      ["A", 400],
      ["B", 300],
      ["C", 200],
      ["D", 100],
    ]);
    const result = computeOtherGrouping(data, "value", "category", {
      limit: 3,
    });
    expect(result.hasOther).toBe(false);
    expect(result.visibleData).toHaveLength(4);
  });

  it("sums Other value correctly", () => {
    const data = makeData([
      ["A", 500],
      ["B", 300],
      ["C", 100],
      ["D", 50],
      ["E", 30],
      ["F", 20],
    ]);
    const result = computeOtherGrouping(data, "value", "category", {
      limit: 3,
    });
    expect(result.hasOther).toBe(true);
    const otherSlice = result.visibleData.find(
      (d) => d.category === OTHER_SLICE_LABEL,
    );
    expect(otherSlice?.value).toBe(100);
    expect(result.otherItems).toHaveLength(3);
    expect(result.total).toBe(1000);
  });

  it("handles empty data", () => {
    const result = computeOtherGrouping([], "value", "category", {});
    expect(result.hasOther).toBe(false);
    expect(result.visibleData).toHaveLength(0);
    expect(result.total).toBe(0);
  });

  it("handles single item", () => {
    const data = makeData([["A", 100]]);
    const result = computeOtherGrouping(data, "value", "category", {});
    expect(result.hasOther).toBe(false);
    expect(result.visibleData).toHaveLength(1);
  });

  it("uses grandTotal when provided for Other value calculation", () => {
    const data = makeData([
      ["A", 500],
      ["B", 300],
      ["C", 100],
      ["D", 50],
      ["E", 30],
      ["F", 20],
    ]);
    const result = computeOtherGrouping(data, "value", "category", {
      limit: 3,
      grandTotal: 2000,
    });
    expect(result.hasOther).toBe(true);
    expect(result.total).toBe(2000);
    const otherSlice = result.visibleData.find(
      (d) => d.category === OTHER_SLICE_LABEL,
    );
    expect(otherSlice?.value).toBe(2000 - 500 - 300 - 100);
  });
});

describe("getOtherTooltipData", () => {
  it("returns top 5 items with percentages", () => {
    const items = makeData(
      Array.from({ length: 8 }, (_, i) => [`Item${i}`, 10 - i]),
    );
    const result = getOtherTooltipData(items, "value", "category", 1000);
    expect(result.items).toHaveLength(5);
    expect(result.remainingCount).toBe(3);
    expect(result.items[0].name).toBe("Item0");
    expect(result.items[0].value).toBe(10);
    expect(result.items[0].percent).toBeCloseTo(1.0, 1);
  });

  it("returns all items when fewer than 5", () => {
    const items = makeData([
      ["X", 20],
      ["Y", 10],
    ]);
    const result = getOtherTooltipData(items, "value", "category", 100);
    expect(result.items).toHaveLength(2);
    expect(result.remainingCount).toBe(0);
    expect(result.totalPercent).toBeCloseTo(30, 1);
  });

  it("calculates total value and percent", () => {
    const items = makeData([
      ["A", 30],
      ["B", 20],
      ["C", 10],
    ]);
    const result = getOtherTooltipData(items, "value", "category", 200);
    expect(result.totalValue).toBe(60);
    expect(result.totalPercent).toBeCloseTo(30, 1);
  });
});
