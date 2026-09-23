import {
  MeasureKeyField,
  SortOrderField,
} from "@rilldata/web-common/features/components/charts/comparison-builder";
import type { TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";
import type { CartesianChartSpec } from "../CartesianChartProvider";
import { generateVLStackedBarChartSpec } from "../stacked-bar/default";
import { at, chartData, stubCanvasContext } from "../test-fixtures";
import { generateVLBarChartSpec } from "./spec";

const base: CartesianChartSpec = {
  metrics_view: "reddit_posts",
  color: "primary",
  x: { field: "post_title", type: "nominal", sort: "-y", limit: 10 },
  y: { field: "post_count", type: "quantitative", zeroBasedOrigin: true },
};

// The same chart with the measure on x and the dimension on y.
const horizontal: CartesianChartSpec = {
  ...base,
  x: { field: "post_count", type: "quantitative", zeroBasedOrigin: true },
  y: { field: "post_title", type: "nominal", sort: "-x", limit: 10 },
};

function paramNames(spec: unknown, path: string): string[] {
  const params = at(spec, path) as { name: string }[];
  return params.map((p) => p.name);
}

describe("generateVLBarChartSpec orientation", () => {
  it("draws the category on y and the measure on x when the measure is on x", () => {
    const spec = generateVLBarChartSpec(horizontal, chartData());

    expect(at(spec, "width")).toBe("container");
    expect(at(spec, "encoding.x")).toBeUndefined();
    expect(at(spec, "encoding.y")).toMatchObject({
      field: "post_title",
      type: "nominal",
      bandPosition: 0,
    });

    expect(at(spec, "layer.0.encoding.y.field")).toBe("post_title");
    expect(at(spec, "layer.0.encoding.x")).toBeUndefined();
    expect(at(spec, "layer.0.params.0.select.encodings")).toEqual(["y"]);

    expect(at(spec, "layer.1.encoding.x")).toMatchObject({
      field: "post_count",
      type: "quantitative",
    });
    expect(at(spec, "layer.1.encoding.y")).toBeUndefined();
    expect(at(spec, "layer.1.mark.height")).toEqual({ band: 0.9 });
    expect(at(spec, "layer.1.mark.width")).toBeUndefined();
  });

  it("sorts the categories by the measure with a channel-relative sort", () => {
    // Without domain values the YAML sort reaches Vega-Lite as is.
    const data = chartData();
    data.domainValues = {};
    const spec = generateVLBarChartSpec(horizontal, data);
    expect(at(spec, "encoding.y.sort")).toBe("-x");
  });

  it("lays out category labels for the axis they end up on", () => {
    // Vertical: automatic x-axis label rotation via signal expressions.
    const vertical = generateVLBarChartSpec(base, chartData());
    expect(at(vertical, "encoding.x.axis.labelAngle")).toHaveProperty("expr");

    // Horizontal: the categories sit on y, where labels stay upright.
    const spec = generateVLBarChartSpec(horizontal, chartData());
    expect(at(spec, "encoding.y.axis.labelAngle")).toBeUndefined();
    expect(at(spec, "encoding.y.axis.labelLimit")).toBeUndefined();
  });

  it("keeps the grid on the measure axis", () => {
    const spec = generateVLBarChartSpec(horizontal, chartData());
    expect(at(spec, "layer.1.encoding.x.axis.grid")).toBe(true);
  });

  it("keeps the axis placement the YAML asked for", () => {
    const spec = generateVLBarChartSpec(
      {
        ...horizontal,
        x: { ...horizontal.x!, axisOrient: "top" },
        y: { ...horizontal.y!, axisOrient: "right" },
      },
      chartData(),
    );
    expect(at(spec, "layer.1.encoding.x.axis.orient")).toBe("top");
    expect(at(spec, "encoding.y.axis.orient")).toBe("right");
  });

  it("groups a color dimension along yOffset", () => {
    const spec = generateVLBarChartSpec(
      { ...horizontal, color: { field: "region", type: "nominal" } },
      chartData({ colorValues: ["us", "eu"] }),
    );

    expect(at(spec, "layer.1.encoding.xOffset")).toBeUndefined();
    expect(at(spec, "layer.1.encoding.yOffset.field")).toBe("region");
  });

  it("offsets comparison periods along yOffset", () => {
    const spec = generateVLBarChartSpec(
      horizontal,
      chartData({ hasComparison: true }),
    );

    expect(at(spec, "layer.1.encoding.xOffset")).toBeUndefined();
    expect(at(spec, "layer.1.encoding.yOffset")).toEqual({
      field: MeasureKeyField,
      sort: { field: SortOrderField },
    });
    expect(at(spec, "layer.1.encoding.opacity")).toBeDefined();
  });

  it("disables brushing for horizontal charts", () => {
    const vertical = generateVLBarChartSpec(
      { ...base, isInteractive: true },
      chartData(),
    );
    expect(at(vertical, "usermeta")).toEqual({
      brushTemporalField: "post_title",
    });
    expect(paramNames(vertical, "layer.0.params")).toContain("brush");

    const spec = generateVLBarChartSpec(
      { ...horizontal, isInteractive: true },
      chartData(),
    );
    expect(at(spec, "usermeta")).toBeUndefined();
    expect(paramNames(spec, "layer.0.params")).not.toContain("brush");
  });

  it("compiles to a valid Vega spec", async () => {
    stubCanvasContext();
    const { compile } = await import("vega-lite");

    const spec = generateVLBarChartSpec(
      { ...horizontal, color: { field: "region", type: "nominal" } },
      chartData({ colorValues: ["us", "eu"], hasComparison: true }),
    );
    expect(() => compile(spec as TopLevelSpec)).not.toThrow();
  });

  it("sizes the height from the container instead of the band step", async () => {
    stubCanvasContext();
    const { compile } = await import("vega-lite");

    const spec = generateVLBarChartSpec(horizontal, chartData());
    expect(at(spec, "height")).toBe("container");

    const { spec: vega } = compile(spec as TopLevelSpec);
    const height = vega.signals?.find((s) => s.name === "height") as
      | { init?: string; update?: string }
      | undefined;
    expect(height?.init).toContain("containerSize()[1]");
    expect(height?.update).toBeUndefined();
  });
});

describe("generateVLStackedBarChartSpec orientation", () => {
  it("stacks along the x channel when the measure is on x", () => {
    const spec = generateVLStackedBarChartSpec(
      { ...horizontal, color: { field: "region", type: "nominal" } },
      chartData({ colorValues: ["us", "eu"] }),
    );

    expect(at(spec, "encoding.y.field")).toBe("post_title");
    expect(at(spec, "layer.1.encoding.x.field")).toBe("post_count");
    expect(at(spec, "layer.1.encoding.y")).toBeUndefined();
    expect(at(spec, "layer.1.encoding.yOffset")).toBeUndefined();
    expect(at(spec, "layer.1.mark.height")).toEqual({ band: 0.9 });
    expect(at(spec, "layer.0.params.0.select.encodings")).toEqual(["y"]);
  });

  it("offsets comparison periods along yOffset", () => {
    const spec = generateVLStackedBarChartSpec(
      horizontal,
      chartData({ hasComparison: true }),
    );
    expect(at(spec, "layer.1.encoding.yOffset.field")).toBe(MeasureKeyField);
  });

  it("compiles to a valid Vega spec", async () => {
    stubCanvasContext();
    const { compile } = await import("vega-lite");

    const spec = generateVLStackedBarChartSpec(
      { ...horizontal, color: { field: "region", type: "nominal" } },
      chartData({ colorValues: ["us", "eu"] }),
    );
    expect(() => compile(spec as TopLevelSpec)).not.toThrow();
  });
});
