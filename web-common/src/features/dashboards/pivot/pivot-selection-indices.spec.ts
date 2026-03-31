import { describe, expect, it } from "vitest";
import {
  isHeaderInHoveredRange,
  isHoveredHeader,
  isInSelectedColRange,
  isInCellSelectedColRange,
} from "./pivot-selection-indices";

// ---- isHeaderInHoveredRange ----

describe("isHeaderInHoveredRange", () => {
  it("returns true when header is within hovered range", () => {
    expect(isHeaderInHoveredRange(2, 1, { start: 0, size: 4 })).toBe(true);
  });

  it("returns true when header exactly matches hovered range", () => {
    expect(isHeaderInHoveredRange(0, 4, { start: 0, size: 4 })).toBe(true);
  });

  it("returns false when header extends beyond hovered range", () => {
    expect(isHeaderInHoveredRange(2, 3, { start: 0, size: 4 })).toBe(false);
  });

  it("returns false when header is before hovered range", () => {
    expect(isHeaderInHoveredRange(0, 1, { start: 2, size: 2 })).toBe(false);
  });

  it("returns false when no hover", () => {
    expect(isHeaderInHoveredRange(0, 1, null)).toBe(false);
  });
});

// ---- isHoveredHeader ----

describe("isHoveredHeader", () => {
  it("returns true when header is the exact hovered header", () => {
    expect(isHoveredHeader(2, 3, { start: 2, size: 3 })).toBe(true);
  });

  it("returns false when start differs", () => {
    expect(isHoveredHeader(0, 3, { start: 2, size: 3 })).toBe(false);
  });

  it("returns false when size differs", () => {
    expect(isHoveredHeader(2, 1, { start: 2, size: 3 })).toBe(false);
  });

  it("returns false when no hover", () => {
    expect(isHoveredHeader(0, 1, null)).toBe(false);
  });
});

// ---- isInSelectedColRange ----

describe("isInSelectedColRange", () => {
  it("returns true when all columns in range are selected", () => {
    const indices = new Set([0, 1, 2, 3]);
    expect(isInSelectedColRange(1, 2, false, indices)).toBe(true);
  });

  it("returns false when not all columns in range are selected", () => {
    const indices = new Set([0, 2]);
    expect(isInSelectedColRange(0, 3, false, indices)).toBe(false);
  });

  it("returns false when self-selected (avoid double-highlighting)", () => {
    const indices = new Set([0, 1]);
    expect(isInSelectedColRange(0, 2, true, indices)).toBe(false);
  });

  it("returns false when no selections", () => {
    expect(isInSelectedColRange(0, 2, false, new Set())).toBe(false);
  });

  it("returns false when colSpan is 0", () => {
    const indices = new Set([0]);
    expect(isInSelectedColRange(0, 0, false, indices)).toBe(false);
  });
});

// ---- isInCellSelectedColRange ----

describe("isInCellSelectedColRange", () => {
  it("returns true when any column in range has a cell selected", () => {
    const indices = new Set([3]);
    expect(isInCellSelectedColRange(2, 3, indices)).toBe(true);
  });

  it("returns false when no columns in range have a cell selected", () => {
    const indices = new Set([5]);
    expect(isInCellSelectedColRange(0, 3, indices)).toBe(false);
  });

  it("returns false with empty set", () => {
    expect(isInCellSelectedColRange(0, 3, new Set())).toBe(false);
  });
});
