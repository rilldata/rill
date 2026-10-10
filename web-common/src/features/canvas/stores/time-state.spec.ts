import { describe, expect, it } from "vitest";
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import { getDefaultComparisonRange } from "./time-state";

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
