import { describe, it, expect } from "vitest";
import { createTDDCartesianSpec } from "./tdd-chart-config";

describe("createTDDCartesianSpec", () => {
  it("sets color to 'primary' when no dimension comparison is active", () => {
    const spec = createTDDCartesianSpec(
      "my_metrics_view",
      "total_sales",
      "timestamp",
    );

    // Without dimension comparison, color should default to "primary"
    // so the chart uses the theme's primary color instead of Vega-Lite's default cyan
    expect(spec.color).toBe("primary");
  });

  it("sets color to a FieldConfig when dimension comparison is active", () => {
    const spec = createTDDCartesianSpec(
      "my_metrics_view",
      "total_sales",
      "timestamp",
      "country",
      ["US", "UK", "DE"],
    );

    expect(spec.color).toEqual({
      field: "country",
      type: "nominal",
      legendOrientation: "none",
      values: ["US", "UK", "DE"],
    });
  });

  it("includes colorMapping from dimensionData when available", () => {
    const dimensionData = [
      { dimensionValue: "US", color: "#ff0000", isFetching: false, data: [] },
      { dimensionValue: "UK", color: "#00ff00", isFetching: false, data: [] },
    ];

    const spec = createTDDCartesianSpec(
      "my_metrics_view",
      "total_sales",
      "timestamp",
      "country",
      ["US", "UK"],
      dimensionData,
    );

    expect(spec.color).toEqual(
      expect.objectContaining({
        field: "country",
        colorMapping: [
          { value: "US", color: "#ff0000" },
          { value: "UK", color: "#00ff00" },
        ],
      }),
    );
  });
});
