import { ColorWithComparisonField } from "@rilldata/web-common/features/components/charts/comparison-builder";
import { describe, expect, it } from "vitest";
import { createVegaTransformPivotConfig } from "./util";

describe("createVegaTransformPivotConfig", () => {
  it("escapes a dotted color dimension so Vega reads it as one field", () => {
    const pivot = createVegaTransformPivotConfig(
      "__time",
      "count",
      "query.context.metaData.type",
      false,
      true,
    );
    expect(pivot).toEqual({
      field: "query\\.context\\.metaData\\.type",
      value: "count",
      groupby: ["__time"],
    });
  });

  it("pivots on the synthetic color field in comparison mode", () => {
    const pivot = createVegaTransformPivotConfig(
      "__time",
      "count",
      "query.context.metaData.type",
      true,
      true,
    );
    expect(pivot?.field).toBe(ColorWithComparisonField);
  });

  it("returns no pivot without a multi-value tooltip", () => {
    expect(
      createVegaTransformPivotConfig("__time", "count", "region", false, false),
    ).toBeUndefined();
  });
});
