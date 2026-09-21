import type { TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";
import type { CartesianChartSpec } from "../CartesianChartProvider";
import { at, chartData, stubCanvasContext } from "../test-fixtures";
import { generateVLStackedBarNormalizedSpec } from "./normalized";

const base: CartesianChartSpec = {
  metrics_view: "reddit_posts",
  color: { field: "region", type: "nominal" },
  x: { field: "post_title", type: "nominal", sort: "-y" },
  y: { field: "post_count", type: "quantitative", zeroBasedOrigin: true },
};

describe("generateVLStackedBarNormalizedSpec orientation", () => {
  it("leaves the vertical layout untouched by default", () => {
    const spec = generateVLStackedBarNormalizedSpec(
      base,
      chartData({ colorValues: ["us", "eu"] }),
    );
    expect(at(spec, "encoding.x.field")).toBe("post_title");
    expect(at(spec, "layer.1.encoding.y.axis.format")).toBe(".0%");
  });

  it("puts the normalized measure axis on x when horizontal", () => {
    const spec = generateVLStackedBarNormalizedSpec(
      { ...base, orientation: "horizontal" },
      chartData({ colorValues: ["us", "eu"] }),
    );

    expect(at(spec, "encoding.y.field")).toBe("post_title");
    expect(at(spec, "encoding.x")).toBeUndefined();

    expect(at(spec, "layer.1.encoding.x.field")).toBe("post_count");
    expect(at(spec, "layer.1.encoding.x.stack")).toBe("normalize");
    expect(at(spec, "layer.1.encoding.x.scale.domainMax")).toBe(1.1);
    expect(at(spec, "layer.1.encoding.x.axis.format")).toBe(".0%");
    expect(at(spec, "layer.1.encoding.y")).toBeUndefined();
    expect(at(spec, "layer.1.mark.height")).toEqual({ band: 0.9 });
    expect(at(spec, "layer.0.params.0.select.encodings")).toEqual(["y"]);
  });

  it("compiles to a valid Vega spec", async () => {
    stubCanvasContext();
    const { compile } = await import("vega-lite");

    const spec = generateVLStackedBarNormalizedSpec(
      { ...base, orientation: "horizontal" },
      chartData({ colorValues: ["us", "eu"], hasComparison: true }),
    );
    expect(() => compile(spec as TopLevelSpec)).not.toThrow();
  });
});
