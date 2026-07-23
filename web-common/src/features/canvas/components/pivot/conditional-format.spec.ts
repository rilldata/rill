import { describe, expect, it } from "vitest";
import {
  conditionalFormatSpecToMeasureFormatting,
  measureFormattingToConditionalFormatSpec,
  type PivotConditionalFormatSpec,
} from "./index";

describe("conditionalFormatSpecToMeasureFormatting", () => {
  it("maps heatmap, data bar and rule specs per measure", () => {
    const specs: PivotConditionalFormatSpec[] = [
      { measure: "revenue", mode: "heatmap", scheme: "red-green" },
      { measure: "cost", mode: "data_bar" },
      {
        measure: "margin",
        mode: "rules",
        rules: [{ operator: "gt", value: 0, color: "green" }],
      },
    ];
    expect(conditionalFormatSpecToMeasureFormatting(specs)).toEqual({
      revenue: { mode: "heatmap", scheme: "red-green" },
      cost: { mode: "data_bar", scheme: "theme-sequential" },
      margin: {
        mode: "rules",
        rules: [{ operator: "gt", value: 0, color: "green" }],
      },
    });
  });

  // The canvas YAML is hand-editable, so malformed values must be skipped
  // instead of crashing the canvas.
  it("returns empty formatting when the field is not an array", () => {
    for (const bad of [undefined, null, 5, "heatmap", { measure: "revenue" }]) {
      expect(conditionalFormatSpecToMeasureFormatting(bad as never)).toEqual(
        {},
      );
    }
  });

  it("skips entries that are not objects or lack a measure name", () => {
    const specs = [
      null,
      "revenue",
      { mode: "heatmap" },
      { measure: "revenue", mode: "heatmap" },
    ];
    expect(conditionalFormatSpecToMeasureFormatting(specs as never)).toEqual({
      revenue: { mode: "heatmap", scheme: "theme-sequential" },
    });
  });

  it("skips rules specs whose rules are not a non-empty array of objects", () => {
    const specs = [
      { measure: "a", mode: "rules", rules: "gt 5" },
      { measure: "b", mode: "rules", rules: [] },
      { measure: "c", mode: "rules", rules: [null, "x"] },
      {
        measure: "d",
        mode: "rules",
        rules: [null, { operator: "gt", value: 0, color: "green" }],
      },
    ];
    expect(conditionalFormatSpecToMeasureFormatting(specs as never)).toEqual({
      d: {
        mode: "rules",
        rules: [{ operator: "gt", value: 0, color: "green" }],
      },
    });
  });

  it("falls back to the default scheme for non-string schemes", () => {
    const specs = [{ measure: "a", mode: "heatmap", scheme: 3 }];
    expect(conditionalFormatSpecToMeasureFormatting(specs as never)).toEqual({
      a: { mode: "heatmap", scheme: "theme-sequential" },
    });
  });

  it("round-trips through measureFormattingToConditionalFormatSpec", () => {
    const specs: PivotConditionalFormatSpec[] = [
      { measure: "revenue", mode: "heatmap", scheme: "red-green" },
      {
        measure: "margin",
        mode: "rules",
        rules: [{ operator: "between", value: 0, value2: 10, color: "blue" }],
      },
    ];
    expect(
      measureFormattingToConditionalFormatSpec(
        conditionalFormatSpecToMeasureFormatting(specs),
      ),
    ).toEqual(specs);
  });
});
