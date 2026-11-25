import * as vega from "vega";
import { compile, type TopLevelSpec } from "vega-lite";
import { describe, expect, it, vi } from "vitest";
import { stubCanvasContext } from "../../features/components/charts/cartesian/test-fixtures";
import { refitOnPlotResize } from "./view-refit";

const rows = Array.from({ length: 24 }, (_, i) => `Series ${i + 1}`).flatMap(
  (series, i) =>
    ["a", "b", "c"].map((category, c) => ({
      series,
      category,
      value: 100 + i * 7 + c,
    })),
);

/** A bar chart of 24 series, fitted to its view, with its legend above the plot. */
function barChart(legend: Record<string, unknown>): TopLevelSpec {
  return {
    width: 560,
    height: 300,
    autosize: { type: "fit", contains: "padding" },
    data: { values: rows },
    mark: "bar",
    encoding: {
      x: { field: "category", type: "nominal" },
      y: { field: "value", type: "quantitative" },
      color: {
        field: "series",
        type: "nominal",
        legend: { orient: "top", ...legend },
      },
    },
  };
}

function headlessView(spec: TopLevelSpec) {
  stubCanvasContext();
  return new vega.View(vega.parse(compile(spec).spec), { renderer: "none" });
}

type SceneItem = {
  role?: string;
  items?: SceneItem[];
  bounds?: { y1: number };
};

/** How far the legend's top edge stands above the top of the view, in pixels. */
function legendOverhang(view: vega.View): number {
  let legend: SceneItem | undefined;
  const visit = (node: SceneItem) => {
    if (node.role === "legend") legend = node.items?.[0];
    for (const item of node.items ?? []) item.items?.forEach(visit);
  };
  visit((view.scenegraph() as unknown as { root: SceneItem }).root);
  if (!legend?.bounds) throw new Error("no legend drawn");
  return -legend.bounds.y1 - view.origin()[1];
}

describe("refitOnPlotResize", () => {
  // The legend takes four columns across the view's 560px. The fit then narrows the plot by the
  // y axis, which leaves room for three, and the legend's two extra rows have nowhere to go.
  const wrappedByWidth = { columns: { expr: "floor(width / 140)" } };

  it("fits the plot around a legend that wraps as the fit narrows the plot", async () => {
    const unfitted = headlessView(barChart(wrappedByWidth));
    unfitted.width(560).height(300);
    await unfitted.runAsync();
    // Vega alone leaves the legend's first rows above the chart.
    expect(legendOverhang(unfitted)).toBeGreaterThan(10);

    const view = headlessView(barChart(wrappedByWidth));
    refitOnPlotResize(view);
    view.width(560).height(300);
    await view.runAsync();
    expect(legendOverhang(view)).toBeLessThanOrEqual(0);
  });

  it("follows a later resize of the view", async () => {
    const view = headlessView(barChart(wrappedByWidth));
    refitOnPlotResize(view);
    view.width(560).height(300);
    await view.runAsync();

    view.width(420).height(300);
    await view.runAsync();
    expect(legendOverhang(view)).toBeLessThanOrEqual(0);
  });

  it("keeps fitting the plot after a run of the view fails", async () => {
    let failing = false;
    vega.expressionFunction("rill_test_fail_while_failing", (width: number) => {
      if (failing) throw new Error("the run fails");
      return width;
    });
    const spec = compile(barChart(wrappedByWidth)).spec;
    spec.signals = [
      ...(spec.signals ?? []),
      {
        name: "rill_test_failure",
        update: "rill_test_fail_while_failing(width)",
      },
    ];
    stubCanvasContext();
    const view = new vega.View(vega.parse(spec), { renderer: "none" });
    view.logLevel(vega.None);
    refitOnPlotResize(view);
    view.width(560).height(300);
    await view.runAsync();

    // Vega drops what was queued to run after a run that fails, the refit among it.
    failing = true;
    view.width(500);
    await view.runAsync();
    failing = false;

    view.width(420);
    await view.runAsync();
    expect(legendOverhang(view)).toBeLessThanOrEqual(0);
  });

  it("gives up on a layout that never settles", async () => {
    // A legend that lists more entries the more room the plot has: the entries take the room,
    // so the plot's height never holds.
    const view = headlessView(
      barChart({
        columns: 1,
        symbolLimit: { expr: "height > 150 ? 24 : 3" },
      }),
    );
    const resize = vi.spyOn(view, "resize");
    refitOnPlotResize(view);
    view.width(560).height(300);
    await view.runAsync();

    // One change of size is followed by a fixed number of fits, not by one per change.
    expect(resize.mock.calls.length).toBeGreaterThan(0);
    expect(resize.mock.calls.length).toBeLessThanOrEqual(3);
  });
});
