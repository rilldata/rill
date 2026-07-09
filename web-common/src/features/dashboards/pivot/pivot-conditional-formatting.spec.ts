import { describe, expect, it } from "vitest";
import {
  computeMeasureDomains,
  getSchemeStops,
  makeCellFormatter,
} from "./pivot-conditional-formatting";

describe("computeMeasureDomains", () => {
  it("computes per-measure min/max independently", () => {
    const domains = computeMeasureDomains([
      { measureName: "revenue", value: 10 },
      { measureName: "revenue", value: 30 },
      { measureName: "revenue", value: 20 },
      { measureName: "orders", value: 1 },
      { measureName: "orders", value: 5 },
    ]);
    expect(domains.get("revenue")).toEqual([10, 30]);
    expect(domains.get("orders")).toEqual([1, 5]);
  });

  it("skips non-finite values and omits measures with no finite values", () => {
    const domains = computeMeasureDomains([
      { measureName: "a", value: NaN },
      { measureName: "a", value: Infinity },
      { measureName: "b", value: -2 },
      { measureName: "b", value: 4 },
    ]);
    expect(domains.has("a")).toBe(false);
    expect(domains.get("b")).toEqual([-2, 4]);
  });
});

describe("getSchemeStops", () => {
  it("returns the scheme stops and reverses when requested", () => {
    const stops = getSchemeStops("greens");
    expect(stops.length).toBeGreaterThan(1);
    expect(getSchemeStops("greens", true)).toEqual([...stops].reverse());
  });

  it("falls back to the default scheme for unknown keys", () => {
    expect(getSchemeStops("does-not-exist")).toBeInstanceOf(Array);
  });
});

describe("makeCellFormatter — heatmap", () => {
  it("maps domain endpoints to the scheme's end colors", () => {
    const fmt = makeCellFormatter([0, 100], {
      mode: "heatmap",
      scheme: "greens",
    });
    const stops = getSchemeStops("greens");
    expect(fmt.background(0).toLowerCase()).toBe(stops[0].toLowerCase());
    expect(fmt.background(100).toLowerCase()).toBe(
      stops[stops.length - 1].toLowerCase(),
    );
  });

  it("uses light text on dark backgrounds and inherits on light", () => {
    const fmt = makeCellFormatter([0, 100], {
      mode: "heatmap",
      scheme: "greens",
    });
    expect(fmt.textColor(0)).toBe("inherit"); // light end
    expect(fmt.textColor(100)).toBe("#ffffff"); // dark end
  });

  it("does not throw on a degenerate domain", () => {
    const fmt = makeCellFormatter([5, 5], { mode: "heatmap", scheme: "blues" });
    expect(() => fmt.background(5)).not.toThrow();
  });
});

describe("makeCellFormatter — data bar", () => {
  it("produces a gradient whose length is proportional to magnitude", () => {
    const fmt = makeCellFormatter([0, 200], {
      mode: "data_bar",
      scheme: "blues",
    });
    expect(fmt.background(0)).toContain("0%");
    expect(fmt.background(100)).toContain("50%");
    expect(fmt.background(200)).toContain("100%");
    expect(fmt.textColor(100)).toBe("inherit");
  });

  it("clamps to 100% and uses magnitude for negative values", () => {
    const fmt = makeCellFormatter([-100, 50], {
      mode: "data_bar",
      scheme: "blues",
    });
    // absMax is 100, so -100 fills the full bar.
    expect(fmt.background(-100)).toContain("100%");
  });
});
