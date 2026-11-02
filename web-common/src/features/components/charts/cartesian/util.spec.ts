import { ColorWithComparisonField } from "@rilldata/web-common/features/components/charts/comparison-builder";
import { describe, expect, it } from "vitest";
import { createVegaTransformPivotConfig } from "./util";

describe("createVegaTransformPivotConfig", () => {
  it("escapes a dotted color field so Vega reads it as a flat field name", () => {
    expect(
      createVegaTransformPivotConfig(
        "__time",
        "count",
        "query.context.type",
        false,
        true,
      ),
    ).toEqual({
      field: "query\\.context\\.type",
      value: "count",
      groupby: ["__time"],
    });
  });

  it("pivots on the combined color field in comparison mode", () => {
    expect(
      createVegaTransformPivotConfig(
        "__time",
        "count",
        "query.context.type",
        true,
        true,
      )?.field,
    ).toBe(ColorWithComparisonField);
  });
});
