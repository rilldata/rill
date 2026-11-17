import { describe, expect, it } from "vitest";
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import {
  getDefaultComparisonRange,
  metricsSQLUsesTimeRange,
} from "./time-state";

describe("getDefaultComparisonRange", () => {
  it("opens with the first comparison declared for the default range", () => {
    const timeRanges = [
      {
        range: "rill-MTD",
        comparisonTimeRanges: [{ offset: "rill-PM" }, { offset: "rill-PY" }],
      },
    ];
    expect(getDefaultComparisonRange("rill-MTD", timeRanges)).toBe(
      TimeComparisonOption.MONTH,
    );
  });

  it("skips declared offsets the canvas URL state cannot use", () => {
    const timeRanges = [
      {
        range: "rill-MTD",
        comparisonTimeRanges: [{ offset: "P1M" }, { offset: "rill-PY" }],
      },
    ];
    expect(getDefaultComparisonRange("rill-MTD", timeRanges)).toBe(
      TimeComparisonOption.YEAR,
    );
  });

  it("ignores comparisons declared for other ranges", () => {
    const timeRanges = [
      { range: "P12M", comparisonTimeRanges: [{ offset: "rill-PY" }] },
    ];
    expect(getDefaultComparisonRange("rill-MTD", timeRanges)).toBe(
      TimeComparisonOption.DAY,
    );
  });

  it("uses the first time range's comparisons when no default range is set", () => {
    const timeRanges = [
      { range: "rill-MTD", comparisonTimeRanges: [{ offset: "rill-PM" }] },
      { range: "P12M", comparisonTimeRanges: [{ offset: "rill-PY" }] },
    ];
    expect(getDefaultComparisonRange(undefined, timeRanges)).toBe(
      TimeComparisonOption.MONTH,
    );
    expect(getDefaultComparisonRange(undefined, [{ range: "rill-MTD" }])).toBe(
      TimeComparisonOption.CONTIGUOUS,
    );
  });

  it("keeps the grain-based default when nothing is declared", () => {
    expect(getDefaultComparisonRange("rill-MTD", undefined)).toBe(
      TimeComparisonOption.DAY,
    );
    expect(getDefaultComparisonRange("rill-QTD", [])).toBe(
      TimeComparisonOption.WEEK,
    );
    expect(getDefaultComparisonRange("P7D", undefined)).toBe(
      TimeComparisonOption.CONTIGUOUS,
    );
    expect(getDefaultComparisonRange(undefined, undefined)).toBe(
      TimeComparisonOption.CONTIGUOUS,
    );
  });
});

describe("metricsSQLUsesTimeRange", () => {
  const timeless = new Map([["products", false]]);
  const mixed = new Map([
    ["products", false],
    ["orders", true],
  ]);

  it.each([
    {
      name: "a canvas without a time dimension never resolves a time range",
      views: ["products"],
      canvas: timeless,
      expected: false,
    },
    {
      name: "a metrics view without a time dimension is not filtered by another one's",
      views: ["products"],
      canvas: mixed,
      expected: false,
    },
    {
      name: "one query on a metrics view with a time dimension is enough",
      views: ["products", "orders"],
      canvas: mixed,
      expected: true,
    },
    {
      name: "an unresolved metrics view follows the canvas when it has a time dimension",
      views: [undefined],
      canvas: mixed,
      expected: true,
    },
    {
      name: "an unresolved metrics view does not wait on a canvas without a time dimension",
      views: [undefined, "unknown"],
      canvas: timeless,
      expected: false,
    },
  ])("$name", ({ views, canvas, expected }) => {
    expect(metricsSQLUsesTimeRange(views, canvas)).toBe(expected);
  });
});
