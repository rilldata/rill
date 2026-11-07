import { describe, expect, it } from "vitest";
import {
  INLINE_CHART_DEFAULT_HEIGHT,
  INLINE_CHART_MAX_HEIGHT,
  INLINE_CHART_MIN_HEIGHT,
  inlineChartHeight,
} from "./sizing";

describe("inlineChartHeight", () => {
  it("defaults when the host sends no dimensions", () => {
    expect(inlineChartHeight(undefined)).toBe(INLINE_CHART_DEFAULT_HEIGHT);
    expect(inlineChartHeight({ width: 600 })).toBe(INLINE_CHART_DEFAULT_HEIGHT);
  });

  it("clamps to the inline range", () => {
    expect(inlineChartHeight({ width: 600, height: 200 })).toBe(
      INLINE_CHART_MIN_HEIGHT,
    );
    expect(inlineChartHeight({ width: 600, maxHeight: 2000 })).toBe(
      INLINE_CHART_MAX_HEIGHT,
    );
    expect(inlineChartHeight({ width: 600, height: 400 }, 100)).toBe(300);
  });
});
