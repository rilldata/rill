import {
  calculateEffectiveRowLimit,
  getNextRowLimit,
  getNextLimitLabel,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import { describe, it, expect } from "vitest";

describe("calculateEffectiveRowLimit", () => {
  describe("with respectPageSize=true (default)", () => {
    it("returns pageSize when rowLimit is undefined", () => {
      expect(calculateEffectiveRowLimit(undefined, 0, 50)).toBe("50");
      expect(calculateEffectiveRowLimit(undefined, 10, 50)).toBe("50");
    });

    it("returns 0 when remainingRows is 0 or negative", () => {
      expect(calculateEffectiveRowLimit(50, 50, 50)).toBe("0");
      expect(calculateEffectiveRowLimit(50, 60, 50)).toBe("0");
    });

    it("returns min of remainingRows and pageSize when both are positive", () => {
      // remainingRows < pageSize
      expect(calculateEffectiveRowLimit(30, 0, 50)).toBe("30");
      expect(calculateEffectiveRowLimit(100, 80, 50)).toBe("20");

      // remainingRows > pageSize
      expect(calculateEffectiveRowLimit(100, 0, 50)).toBe("50");
      expect(calculateEffectiveRowLimit(150, 50, 50)).toBe("50");

      // remainingRows == pageSize
      expect(calculateEffectiveRowLimit(50, 0, 50)).toBe("50");
    });

    it("handles pagination correctly with rowOffset", () => {
      const rowLimit = 100;
      const pageSize = 50;

      // First page
      expect(calculateEffectiveRowLimit(rowLimit, 0, pageSize)).toBe("50");

      // Second page
      expect(calculateEffectiveRowLimit(rowLimit, 50, pageSize)).toBe("50");

      // Third page (only 10 rows left)
      expect(calculateEffectiveRowLimit(rowLimit, 90, pageSize)).toBe("10");
    });
  });

  describe("with respectPageSize=false", () => {
    it("returns pageSize when rowLimit is undefined", () => {
      expect(calculateEffectiveRowLimit(undefined, 0, 50, false)).toBe("50");
    });

    it("returns 0 when remainingRows is 0 or negative", () => {
      expect(calculateEffectiveRowLimit(50, 50, 50, false)).toBe("0");
      expect(calculateEffectiveRowLimit(50, 60, 50, false)).toBe("0");
    });

    it("returns full remainingRows without pageSize constraint", () => {
      // This is the key difference - no min() with pageSize
      expect(calculateEffectiveRowLimit(100, 0, 50, false)).toBe("100");
      expect(calculateEffectiveRowLimit(75, 0, 50, false)).toBe("75");
      expect(calculateEffectiveRowLimit(150, 0, 50, false)).toBe("150");
    });

    it("handles offset correctly without pageSize constraint", () => {
      expect(calculateEffectiveRowLimit(100, 20, 50, false)).toBe("80");
      expect(calculateEffectiveRowLimit(100, 50, 50, false)).toBe("50");
      expect(calculateEffectiveRowLimit(100, 90, 50, false)).toBe("10");
    });

    it("allows fetching more than one page at once", () => {
      // When user clicks "Show more" to 100, we want to fetch all 100 rows
      // not just 50 (one page)
      expect(calculateEffectiveRowLimit(100, 0, 50, false)).toBe("100");
    });
  });
});

describe("getNextRowLimit", () => {
  it("returns next limit in progression: 5 → 10 → 25 → 50 → 100", () => {
    expect(getNextRowLimit(5)).toBe(10);
    expect(getNextRowLimit(10)).toBe(25);
    expect(getNextRowLimit(25)).toBe(50);
    expect(getNextRowLimit(50)).toBe(100);
  });

  it("returns undefined when at or beyond 100", () => {
    expect(getNextRowLimit(100)).toBeUndefined();
    expect(getNextRowLimit(150)).toBeUndefined();
  });

  it("finds next higher value when current limit is not in progression", () => {
    expect(getNextRowLimit(7)).toBe(10);
    expect(getNextRowLimit(15)).toBe(25);
    expect(getNextRowLimit(30)).toBe(50);
    expect(getNextRowLimit(60)).toBe(100);
  });

  it("handles edge cases", () => {
    expect(getNextRowLimit(1)).toBe(5);
    expect(getNextRowLimit(99)).toBe(100);
  });
});

describe("getNextLimitLabel", () => {
  it("returns string representation of next limit", () => {
    expect(getNextLimitLabel(5)).toBe("10");
    expect(getNextLimitLabel(10)).toBe("25");
    expect(getNextLimitLabel(25)).toBe("50");
    expect(getNextLimitLabel(50)).toBe("100");
  });

  it("returns '100' when at or beyond max limit", () => {
    expect(getNextLimitLabel(100)).toBe("100");
    expect(getNextLimitLabel(150)).toBe("100");
  });

  it("handles non-standard limits", () => {
    expect(getNextLimitLabel(7)).toBe("10");
    expect(getNextLimitLabel(30)).toBe("50");
  });
});
