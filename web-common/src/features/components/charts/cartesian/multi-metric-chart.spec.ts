import type { TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";
import type { CartesianChartSpec } from "./CartesianChartProvider";
import { generateVLMultiMetricChartSpec } from "./multi-metric-chart";
import { at, chartData, stubCanvasContext } from "./test-fixtures";

const base: CartesianChartSpec = {
  metrics_view: "reddit_posts",
  color: { field: "rill_measures", type: "value" },
  x: { field: "post_title", type: "nominal", sort: "-y" },
  y: {
    field: "post_count",
    fields: ["post_count", "comment_count"],
    type: "quantitative",
    zeroBasedOrigin: true,
  },
};

// The same chart with the measures on x and the dimension on y.
const horizontal: CartesianChartSpec = {
  ...base,
  x: {
    field: "post_count",
    fields: ["post_count", "comment_count"],
    type: "quantitative",
    zeroBasedOrigin: true,
  },
  y: { field: "post_title", type: "nominal", sort: "-x" },
};

describe("generateVLMultiMetricChartSpec orientation", () => {
  it("draws grouped measures along yOffset when the measures are on x", () => {
    const spec = generateVLMultiMetricChartSpec(
      horizontal,
      chartData(),
      "grouped_bar",
    );

    expect(at(spec, "encoding.y.field")).toBe("post_title");
    expect(at(spec, "encoding.x")).toBeUndefined();

    expect(at(spec, "layer.1.encoding.x.field")).toBe("value");
    expect(at(spec, "layer.1.encoding.y")).toBeUndefined();
    expect(at(spec, "layer.1.encoding.yOffset.field")).toBe("Measure");
    expect(at(spec, "layer.1.encoding.xOffset")).toBeUndefined();
    expect(at(spec, "layer.1.mark.height")).toEqual({ band: 0.9 });
    expect(at(spec, "layer.0.params.0.select.encodings")).toEqual(["y"]);
  });

  it("stacks measures along x and offsets comparison periods on yOffset", () => {
    const spec = generateVLMultiMetricChartSpec(
      horizontal,
      chartData({ hasComparison: true }),
      "stacked_bar",
    );

    expect(at(spec, "layer.1.encoding.x.stack")).toBe("zero");
    expect(at(spec, "layer.1.encoding.yOffset.field")).toBe("period");
  });

  it("puts the normalized axis on x", () => {
    const spec = generateVLMultiMetricChartSpec(
      horizontal,
      chartData(),
      "stacked_bar_normalized",
    );

    expect(at(spec, "layer.1.encoding.x.stack")).toBe("normalize");
    expect(at(spec, "layer.1.encoding.x.axis.format")).toBe(".0%");
  });

  it("renders the same vertical layout for the line and area variants", () => {
    // A measure on x is rejected for line and area charts upstream; the
    // builder still reads the fields by role so nothing breaks.
    for (const markType of ["line", "stacked_area"] as const) {
      expect(
        generateVLMultiMetricChartSpec(horizontal, chartData(), markType),
      ).toEqual(generateVLMultiMetricChartSpec(base, chartData(), markType));
    }
  });

  it("compiles to a valid Vega spec", async () => {
    stubCanvasContext();
    const { compile } = await import("vega-lite");

    for (const markType of [
      "grouped_bar",
      "stacked_bar",
      "stacked_bar_normalized",
    ] as const) {
      const spec = generateVLMultiMetricChartSpec(
        horizontal,
        chartData({ hasComparison: true }),
        markType,
      );
      expect(() => compile(spec as TopLevelSpec)).not.toThrow();
    }
  });
});
