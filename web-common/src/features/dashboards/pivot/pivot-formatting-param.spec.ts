import { describe, expect, it } from "vitest";
import {
  fromPivotFormattingParam,
  toPivotFormattingParam,
} from "./pivot-formatting-param";
import type { PivotMeasureFormatting } from "./types";

describe("pivot formatting url param codec", () => {
  it("serializes scale modes", () => {
    expect(
      toPivotFormattingParam({
        revenue: { mode: "heatmap", scheme: "greens" },
        orders: { mode: "heatmap", scheme: "blues" },
        margin: { mode: "data_bar", scheme: "theme-sequential" },
      }),
    ).toBe(
      "revenue:heatmap:greens;orders:heatmap:blues;margin:data_bar:theme-sequential",
    );
  });

  it("serializes threshold rules, stripping # from hex colors", () => {
    expect(
      toPivotFormattingParam({
        margin: {
          mode: "rules",
          rules: [
            { operator: "lt", value: 0, color: "negative" },
            { operator: "between", value: 0, value2: 0.4, color: "#ffcc00" },
            { operator: "gte", value: 0.4, color: "positive" },
          ],
        },
      }),
    ).toBe("margin:rules:lt,0,negative|between,0,0.4,ffcc00|gte,0.4,positive");
  });

  it("returns an empty string for empty or undefined config", () => {
    expect(toPivotFormattingParam(undefined)).toBe("");
    expect(toPivotFormattingParam({})).toBe("");
  });

  it("round-trips every mode", () => {
    const config: Record<string, PivotMeasureFormatting> = {
      revenue: { mode: "heatmap", scheme: "greens" },
      orders: { mode: "data_bar", scheme: "blues" },
      margin: {
        mode: "rules",
        rules: [
          { operator: "lt", value: 0, color: "negative" },
          { operator: "between", value: 0, value2: 0.4, color: "ffcc00" },
        ],
      },
    };
    const param = toPivotFormattingParam(config);
    const { measureFormatting, invalidEntries } =
      fromPivotFormattingParam(param);
    expect(invalidEntries).toEqual([]);
    expect(measureFormatting).toEqual(config);
    // Stability: re-serializing the parsed value yields the same param.
    expect(toPivotFormattingParam(measureFormatting)).toBe(param);
  });

  it("parses an empty param to undefined", () => {
    expect(fromPivotFormattingParam("").measureFormatting).toBeUndefined();
  });

  it("rejects malformed entries but keeps valid ones", () => {
    const { measureFormatting, invalidEntries } = fromPivotFormattingParam(
      [
        "revenue:heatmap:greens",
        "orders:heatmap", // missing scheme
        "margin:rules:noop,1,red", // unknown operator
        "profit:rules:lt,abc,red", // non-numeric value
        "cost:sparkles:greens", // unknown mode
        "tax:heatmap:blues:reverse", // trailing extra part
        "sales:rules:between,1,2,blue", // valid
      ].join(";"),
    );
    expect(Object.keys(measureFormatting ?? {})).toEqual(["revenue", "sales"]);
    expect(invalidEntries).toHaveLength(5);
  });

  it("rejects rules with a wrong number of parts", () => {
    expect(
      fromPivotFormattingParam("m:rules:lt,0").measureFormatting,
    ).toBeUndefined();
    expect(
      fromPivotFormattingParam("m:rules:lt,0,red,extra").measureFormatting,
    ).toBeUndefined();
    expect(
      fromPivotFormattingParam("m:rules:between,0,1").measureFormatting,
    ).toBeUndefined();
  });
});
