import type { ChartDataResult } from "@rilldata/web-common/features/components/charts/types";
import { generateSpec } from "@rilldata/web-common/features/components/charts/util";
import chroma from "chroma-js";
import { compile } from "vega-lite";
import { describe, expect, it } from "vitest";
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

describe("dynamic y-axis", () => {
  const rows = [
    { timestamp: "2024-01-01", total_sales: 5, country: "US" },
    { timestamp: "2024-01-02", total_sales: 9, country: "UK" },
  ];

  const data = {
    data: rows,
    isFetching: false,
    error: undefined,
    fields: {
      total_sales: { displayName: "Total Sales" },
      timestamp: { displayName: "Time" },
      country: { displayName: "Country" },
    },
    domainValues: { country: ["US", "UK"] },
    theme: { primary: chroma("#4a5ad0"), secondary: chroma("#a0a7e0") },
    isDarkMode: false,
    hasComparison: false,
  } as unknown as ChartDataResult;

  /**
   * Compiles all the way down to Vega, because the y scale's `zero` flag alone does not tell
   * you whether the axis is actually dynamic: Vega-Lite stacks bar and area marks by default,
   * and a stacked domain is derived from start/end fields that always span zero.
   */
  function compiledYScale(dynamicYAxis: boolean, comparisonDimension?: string) {
    const config = createTDDCartesianSpec(
      "my_metrics_view",
      "total_sales",
      "timestamp",
      comparisonDimension,
      comparisonDimension ? ["US", "UK"] : [],
      [],
      false,
      dynamicYAxis,
    );
    const vlSpec = generateSpec("bar_chart", config, data);
    const compiled = compile({ ...vlSpec, data: { values: rows } } as never)
      .spec as {
      scales?: { name: string; zero?: boolean; domain?: unknown }[];
    };
    return compiled.scales?.find((scale) => scale.name === "y");
  }

  it("anchors the axis at zero when the toggle is off", () => {
    const yScale = compiledYScale(false);

    expect(yScale?.zero).toBe(true);
  });

  it("drops stacking so the domain follows the data when the toggle is on", () => {
    const yScale = compiledYScale(true);

    expect(yScale?.zero).toBe(false);
    // A stacked domain would be {fields: [..._start, ..._end]}, which always spans zero.
    expect(yScale?.domain).toEqual({ data: "data_2", field: "total_sales" });
  });

  it("keeps multi-series charts stacked, since un-stacking would overlap the marks", () => {
    const yScale = compiledYScale(true, "country");

    expect(yScale?.domain).toEqual({
      data: "data_2",
      fields: ["total_sales_start", "total_sales_end"],
    });
  });
});
