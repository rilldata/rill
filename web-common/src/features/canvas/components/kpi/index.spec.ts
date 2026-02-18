import { describe, expect, it } from "vitest";
import { getMinWidth, getOptimalColumns } from "./index";

describe("getMinWidth", () => {
  it("returns correct width for sparkline right", () => {
    expect(getMinWidth("right")).toBe(328); // SPARK_RIGHT_MIN
  });

  it("returns correct width for sparkline bottom", () => {
    expect(getMinWidth("bottom")).toBe(192); // BIG_NUMBER_MIN_WIDTH + padding
  });

  it("returns correct width for sparkline none", () => {
    expect(getMinWidth("none")).toBe(192);
  });

  it("returns correct width for undefined", () => {
    expect(getMinWidth(undefined)).toBe(192);
  });
});

describe("getOptimalColumns", () => {
  describe("edge cases", () => {
    it("returns 1 for zero items", () => {
      expect(getOptimalColumns(0, 1000, 200)).toBe(1);
    });

    it("returns 1 for negative items", () => {
      expect(getOptimalColumns(-1, 1000, 200)).toBe(1);
    });

    it("returns 1 for zero container width", () => {
      expect(getOptimalColumns(6, 0, 200)).toBe(1);
    });

    it("returns 1 for negative container width", () => {
      expect(getOptimalColumns(6, -100, 200)).toBe(1);
    });

    it("returns 1 for zero min width", () => {
      expect(getOptimalColumns(6, 1000, 0)).toBe(1);
    });

    it("returns 1 for single item", () => {
      expect(getOptimalColumns(1, 1000, 200)).toBe(1);
    });
  });

  describe("even distribution - avoids whitespace", () => {
    it("returns 3 for 6 items when container fits 4+ columns (3x2 layout)", () => {
      // Container width 800px, minWidth 192px = max 4 columns
      // 6 items: factors are 1, 2, 3, 6
      // Best factor <= 4 is 3 (gives 3x2 layout, no whitespace)
      expect(getOptimalColumns(6, 800, 192)).toBe(3);
    });

    it("returns 2 for 6 items when container fits 2-3 columns (2x3 layout)", () => {
      // Container width 500px, minWidth 192px = max 2 columns
      // 6 items: factors are 1, 2, 3, 6
      // Best factor <= 2 is 2 (gives 2x3 layout, no whitespace)
      expect(getOptimalColumns(6, 500, 192)).toBe(2);
    });

    it("returns 4 for 8 items when container fits 4 columns (4x2 layout)", () => {
      // Container width 800px, minWidth 192px = max 4 columns
      // 8 items: factors are 1, 2, 4, 8
      // Best factor <= 4 is 4 (gives 4x2 layout, no whitespace)
      expect(getOptimalColumns(8, 800, 192)).toBe(4);
    });

    it("returns 2 for 4 items when container fits 3 columns (2x2 layout)", () => {
      // Container width 600px, minWidth 192px = max 3 columns
      // 4 items: factors are 1, 2, 4
      // Best factor <= 3 is 2 (gives 2x2 layout, no whitespace)
      expect(getOptimalColumns(4, 600, 192)).toBe(2);
    });

    it("returns 3 for 9 items when container fits 4 columns (3x3 layout)", () => {
      // 9 items: factors are 1, 3, 9
      // Best factor <= 4 is 3 (gives 3x3 layout, no whitespace)
      expect(getOptimalColumns(9, 800, 192)).toBe(3);
    });

    it("returns 5 for 10 items when container fits 5 columns (5x2 layout)", () => {
      // 10 items: factors are 1, 2, 5, 10
      // Container width 1000px, minWidth 192px = max 5 columns
      // Best factor <= 5 is 5 (gives 5x2 layout, no whitespace)
      expect(getOptimalColumns(10, 1000, 192)).toBe(5);
    });
  });

  describe("prime numbers - minimizes whitespace", () => {
    it("handles 5 items (prime) by minimizing empty cells", () => {
      // Container width 800px, minWidth 192px = max 4 columns
      // 5 is prime, no perfect factors
      // With 4 columns: 2 rows, 3 empty cells
      // With 3 columns: 2 rows, 1 empty cell (better)
      // With 2 columns: 3 rows, 1 empty cell
      // With 5 columns: 1 row, 0 empty cells (best if fits)
      const result = getOptimalColumns(5, 800, 192);
      // Should pick columns that minimize empty cells
      expect(result).toBeGreaterThanOrEqual(1);
      expect(result).toBeLessThanOrEqual(4);
    });

    it("handles 7 items (prime)", () => {
      const result = getOptimalColumns(7, 800, 192);
      expect(result).toBeGreaterThanOrEqual(1);
      expect(result).toBeLessThanOrEqual(4);
    });
  });

  describe("respects container width constraint", () => {
    it("returns 1 when container only fits 1 column", () => {
      // Container width 150px, minWidth 192px = max 0 columns -> 1
      expect(getOptimalColumns(6, 150, 192)).toBe(1);
    });

    it("limits columns to what fits in container", () => {
      // Container width 400px, minWidth 192px = max 2 columns
      // 12 items: would ideally be 4x3, but only 2 columns fit
      expect(getOptimalColumns(12, 400, 192)).toBe(2);
    });
  });

  describe("real-world scenarios", () => {
    it("handles sportradar example: 6 KPIs in wide container", () => {
      // Based on the issue screenshot showing 6 KPIs
      // Should give 3x2 or 2x3 layout, not 4+2
      const columns = getOptimalColumns(6, 1000, 192);
      expect(columns).toBe(3); // 3x2 layout is optimal
    });

    it("handles 3 KPIs - should be 3x1", () => {
      const columns = getOptimalColumns(3, 1000, 192);
      expect(columns).toBe(3);
    });

    it("handles 4 KPIs - should be 4x1 or 2x2", () => {
      const columns = getOptimalColumns(4, 1000, 192);
      expect(columns).toBe(4); // Largest factor that fits
    });
  });
});
