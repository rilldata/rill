import { describe, expect, it } from "vitest";
import {
  computeMeasureDomains,
  getSchemeStops,
  makeCellFormatter,
  PIVOT_RULE_COLORS,
  resolveRuleColor,
  ruleMatches,
} from "./pivot-conditional-formatting";
import type { PivotFormatRule } from "./types";

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
    const fmt = makeCellFormatter(
      { mode: "heatmap", scheme: "greens" },
      [0, 100],
    );
    const stops = getSchemeStops("greens");
    expect(fmt(0)?.background.toLowerCase()).toBe(stops[0].toLowerCase());
    expect(fmt(100)?.background.toLowerCase()).toBe(
      stops[stops.length - 1].toLowerCase(),
    );
  });

  it("uses light text on dark backgrounds and inherits on light", () => {
    const fmt = makeCellFormatter(
      { mode: "heatmap", scheme: "greens" },
      [0, 100],
    );
    expect(fmt(0)?.color).toBe("inherit"); // light end
    expect(fmt(100)?.color).toBe("#ffffff"); // dark end
  });

  it("does not throw on a degenerate domain", () => {
    const fmt = makeCellFormatter({ mode: "heatmap", scheme: "blues" }, [5, 5]);
    expect(() => fmt(5)).not.toThrow();
  });

  it("flips the gradient for lower-is-better measures", () => {
    const fmt = makeCellFormatter(
      { mode: "heatmap", scheme: "greens" },
      [0, 100],
      true,
    );
    const stops = getSchemeStops("greens");
    expect(fmt(0)?.background.toLowerCase()).toBe(
      stops[stops.length - 1].toLowerCase(),
    );
    expect(fmt(100)?.background.toLowerCase()).toBe(stops[0].toLowerCase());
  });
});

describe("makeCellFormatter — data bar", () => {
  it("produces a gradient whose length is proportional to magnitude", () => {
    const fmt = makeCellFormatter(
      { mode: "data_bar", scheme: "blues" },
      [0, 200],
    );
    expect(fmt(0)?.background).toContain("0%");
    expect(fmt(100)?.background).toContain("50%");
    expect(fmt(200)?.background).toContain("100%");
    expect(fmt(100)?.color).toBe("inherit");
  });

  it("clamps to 100% and uses magnitude for negative values", () => {
    const fmt = makeCellFormatter(
      { mode: "data_bar", scheme: "blues" },
      [-100, 50],
    );
    // absMax is 100, so -100 fills the full bar.
    expect(fmt(-100)?.background).toContain("100%");
  });

  it("ignores lower-is-better: bars encode magnitude, not direction", () => {
    const fmt = makeCellFormatter(
      { mode: "data_bar", scheme: "blues" },
      [0, 200],
    );
    const fmtLowerIsBetter = makeCellFormatter(
      { mode: "data_bar", scheme: "blues" },
      [0, 200],
      true,
    );
    expect(fmtLowerIsBetter(100)).toEqual(fmt(100));
  });
});

describe("ruleMatches", () => {
  it("evaluates each operator", () => {
    expect(ruleMatches({ operator: "gt", value: 5, color: "x" }, 6)).toBe(true);
    expect(ruleMatches({ operator: "gt", value: 5, color: "x" }, 5)).toBe(
      false,
    );
    expect(ruleMatches({ operator: "gte", value: 5, color: "x" }, 5)).toBe(
      true,
    );
    expect(ruleMatches({ operator: "lt", value: 5, color: "x" }, 4)).toBe(true);
    expect(ruleMatches({ operator: "lte", value: 5, color: "x" }, 5)).toBe(
      true,
    );
    expect(ruleMatches({ operator: "eq", value: 5, color: "x" }, 5)).toBe(true);
    expect(ruleMatches({ operator: "eq", value: 5, color: "x" }, 5.1)).toBe(
      false,
    );
  });

  it("treats between as inclusive of both bounds", () => {
    const rule: PivotFormatRule = {
      operator: "between",
      value: 0,
      value2: 10,
      color: "x",
    };
    expect(ruleMatches(rule, 0)).toBe(true);
    expect(ruleMatches(rule, 10)).toBe(true);
    expect(ruleMatches(rule, 10.1)).toBe(false);
    expect(ruleMatches(rule, -0.1)).toBe(false);
  });
});

describe("resolveRuleColor", () => {
  it("resolves semantic keys to their paired fill and text colors", () => {
    const positive = PIVOT_RULE_COLORS.find((c) => c.key === "positive")!;
    expect(resolveRuleColor("positive")).toEqual({
      fill: positive.fill,
      text: positive.text,
    });
  });

  it("accepts raw hex with or without the leading #", () => {
    expect(resolveRuleColor("#00441b").fill).toBe("#00441b");
    expect(resolveRuleColor("00441b").fill).toBe("#00441b");
    // Dark fill gets light text.
    expect(resolveRuleColor("00441b").text).toBe("#ffffff");
  });

  it("falls back to neutral for invalid colors", () => {
    const neutral = PIVOT_RULE_COLORS.find((c) => c.key === "neutral")!;
    expect(resolveRuleColor("not-a-color")).toEqual({
      fill: neutral.fill,
      text: neutral.text,
    });
  });
});

describe("makeCellFormatter — rules", () => {
  const rules: PivotFormatRule[] = [
    { operator: "lt", value: 0, color: "negative" },
    { operator: "gte", value: 0.4, color: "positive" },
  ];

  it("applies the first matching rule and leaves other values unformatted", () => {
    const fmt = makeCellFormatter({ mode: "rules", rules });
    const negative = PIVOT_RULE_COLORS.find((c) => c.key === "negative")!;
    const positive = PIVOT_RULE_COLORS.find((c) => c.key === "positive")!;
    expect(fmt(-1)).toEqual({
      background: negative.fill,
      color: negative.text,
    });
    expect(fmt(0.5)).toEqual({
      background: positive.fill,
      color: positive.text,
    });
    expect(fmt(0.2)).toBeNull();
  });

  it("respects rule order: first match wins", () => {
    const fmt = makeCellFormatter({
      mode: "rules",
      rules: [
        { operator: "gt", value: 0, color: "warning" },
        { operator: "gt", value: 10, color: "positive" },
      ],
    });
    const warning = PIVOT_RULE_COLORS.find((c) => c.key === "warning")!;
    // 20 matches both rules; the earlier one wins.
    expect(fmt(20)?.background).toBe(warning.fill);
  });
});
