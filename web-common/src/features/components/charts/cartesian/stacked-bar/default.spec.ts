import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import { ColorWithComparisonField } from "@rilldata/web-common/features/components/charts/comparison-builder";
import type { Scene } from "vega";
import type { TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";
import type { CartesianChartSpec } from "../CartesianChartProvider";
import { at, chartData, stubCanvasContext } from "../test-fixtures";
import { generateVLStackedBarChartSpec } from "./default";

// A nested Druid column surfaces as a dimension whose name contains dots.
// Vega reads an unescaped dotted field as a nested lookup, so every place the
// spec names the dimension has to escape it or the chart renders nothing.
const dimension = "query.context.metaData.type";
const escapedDimension = "query\\.context\\.metaData\\.type";
const values = ["Scheduled Report", "Other"];

const config: CartesianChartSpec = {
  metrics_view: "druid_requests",
  x: { field: "__time", type: "temporal" },
  y: { field: "count", type: "quantitative", zeroBasedOrigin: true },
  color: { field: dimension, type: "nominal", values },
};

function data(hasComparison = false) {
  const result = chartData({ hasComparison });
  result.fields = {
    __time: { name: "__time", displayName: "Time" },
    count: { name: "count", displayName: "Count" },
    [dimension]: { name: dimension, displayName: "Query Type" },
  };
  result.domainValues = { [dimension]: values };
  return result;
}

describe("generateVLStackedBarChartSpec with a dotted color dimension", () => {
  it("escapes the dimension in the pivot and color encodings", () => {
    const spec = generateVLStackedBarChartSpec(config, data());
    expect(at(spec, "layer.0.transform.0.pivot")).toBe(escapedDimension);
    expect(at(spec, "layer.1.encoding.color.field")).toBe(escapedDimension);
  });

  it("pivots on the synthetic color field in comparison mode", () => {
    const spec = generateVLStackedBarChartSpec(config, data(true));
    expect(at(spec, "layer.0.transform.0.pivot")).toBe(
      ColorWithComparisonField,
    );
  });

  it("renders a bar per row when rows are keyed by the dotted dimension", async () => {
    stubCanvasContext();
    const { compile } = await import("vega-lite");
    const vega = await import("vega");
    // The app registers a formatter per measure for tooltips (see Chart.svelte).
    vega.expressionFunction(sanitizeFieldName("count"), (value: unknown) =>
      String(value),
    );

    const spec = generateVLStackedBarChartSpec(config, data());
    const rows = [
      { __time: "2026-10-27T00:00:00Z", count: 2, [dimension]: values[0] },
      { __time: "2026-10-27T00:00:00Z", count: 3, [dimension]: values[1] },
      { __time: "2026-10-27T01:00:00Z", count: 4, [dimension]: values[0] },
      { __time: "2026-10-27T01:00:00Z", count: 1, [dimension]: values[1] },
    ];

    // Vega reports dataflow failures through its logger instead of throwing.
    const errors: unknown[] = [];
    const view = new vega.View(vega.parse(compile(spec as TopLevelSpec).spec), {
      renderer: "none",
    })
      .logger(
        vega.logger(vega.Error, "error", (_method, _level, args) => {
          errors.push(...Array.from(args as ArrayLike<unknown>));
        }),
      )
      .width(400)
      .height(200);
    view.data("metrics-view", rows);
    await view.runAsync();

    expect(errors).toEqual([]);
    // At runtime `scenegraph()` returns the Scenegraph wrapper, not the root mark.
    const { root } = view.scenegraph() as unknown as { root: Scene };
    expect(countBars(root)).toBe(rows.length);
  });
});

/** Counts the painted bar marks, skipping the transparent hover targets. */
function countBars(mark: Scene): number {
  const own =
    mark.marktype === "rect"
      ? mark.items.filter(
          (bar: { fill?: string; height?: number }) =>
            bar.fill !== "transparent" && (bar.height ?? 0) > 0,
        ).length
      : 0;
  return mark.items
    .flatMap((group) => group.items ?? [])
    .reduce((sum, child) => sum + countBars(child), own);
}
