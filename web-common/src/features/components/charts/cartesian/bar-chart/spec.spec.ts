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

const horizontal: CartesianChartSpec = { ...base, orientation: "horizontal" };

function paramNames(spec: unknown, path: string): string[] {
  const params = at(spec, path) as { name: string }[];
  return params.map((p) => p.name);
}

describe("generateVLBarChartSpec orientation", () => {
  it("treats an explicit vertical orientation as the default", () => {
    expect(
      generateVLBarChartSpec({ ...base, orientation: "vertical" }, chartData()),
    ).toEqual(generateVLBarChartSpec(base, chartData()));
  });

  it("draws the category on y and the measure on x when horizontal", () => {
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

  it("keeps the grid on the measure axis", () => {
    const spec = generateVLBarChartSpec(horizontal, chartData());
    expect(at(spec, "layer.1.encoding.x.axis.grid")).toBe(true);
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
});

describe("generateVLStackedBarChartSpec orientation", () => {
  it("stacks along the x channel when horizontal", () => {
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
